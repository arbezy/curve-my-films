package main

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

var templates = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

// serveIndex serves the HTMX-driven frontend's single HTML page.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serveReviewsFragment renders the review list as an HTML fragment for HTMX to swap in.
func serveReviewsFragment(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	revs, err := rr.FetchAllReviews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "reviews.html", revs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// deleteReviewUI handles a review deletion triggered from the review list.
// On success it returns 200 OK with an empty body — the delete button's
// hx-swap="delete" removes the review's <li> from the list regardless of
// response content.
func deleteReviewUI(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	switch err := DeleteReviewByID(&rr, id); {
	case err == nil:
		w.WriteHeader(http.StatusOK)
	case errors.Is(err, ErrReviewNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, fmt.Sprintf("error deleting review: %v", err), http.StatusInternalServerError)
	}
}

// editReviewView is the data for the initial "edit a review" form.
type editReviewView struct {
	ReviewID  int
	MovieName string
	Rating    int
}

// serveEditReviewForm renders the initial edit form for one review, pre-filled
// with its current movie name and rating.
func serveEditReviewForm(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	rev, err := rr.FetchReviewByID(id)
	switch {
	case errors.Is(err, ErrReviewNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := editReviewView{ReviewID: rev.ReviewID, MovieName: rev.MovieName, Rating: rev.Rating}
	if err := templates.ExecuteTemplate(w, "edit_review_form.html", view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// editCompareView is the data for one step of the edit comparison flow — shown
// only once a rating change requires the review to be re-ranked into its new
// rating's tree, mirroring compareView for the add-review flow.
type editCompareView struct {
	ReviewID  int
	MovieName string
	Rating    int
	Path      []string
	Current   *RatingTreeNode
}

// editReviewResultView is the data for the outcome of an edit: a banner plus
// the refreshed review list underneath it either way.
type editReviewResultView struct {
	MovieName string
	Rating    int
	Err       error
	Reviews   []*MovieReview
}

// editReviewCompare handles a step of the edit flow: it's posted to both by the
// initial edit form and by each comparison choice. If the rating hasn't
// changed, the edit is applied immediately — reordering a rating's tree only
// matters when a review moves into a different rating. Otherwise it walks the
// new rating's tree the same way compareReview does, and once the path reaches
// an empty slot, repositions the review there.
func editReviewCompare(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reviewID, err := strconv.Atoi(r.Form.Get("review_id"))
	if err != nil {
		http.Error(w, "review_id must be an integer", http.StatusBadRequest)
		return
	}

	movieName := r.Form.Get("movie_name")
	if len(movieName) == 0 {
		http.Error(w, "movie_name is required", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.Form.Get("rating"))
	if err != nil {
		http.Error(w, "rating must be an integer", http.StatusBadRequest)
		return
	}

	path := r.Form["path"]

	current, err := rr.FetchReviewByID(reviewID)
	if err != nil {
		renderEditReviewResult(w, movieName, rating, err)
		return
	}

	if rating == current.Rating {
		err := rr.UpdateReviewDetails(reviewID, movieName, rating)
		renderEditReviewResult(w, movieName, rating, err)
		return
	}

	node, err := NodeAtPath(&rr, rating, path)
	if err != nil {
		renderEditReviewResult(w, movieName, rating, err)
		return
	}

	if node != nil {
		view := editCompareView{ReviewID: reviewID, MovieName: movieName, Rating: rating, Path: path, Current: node}
		if err := templates.ExecuteTemplate(w, "edit_compare.html", view); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = RepositionReviewAtPath(&rr, reviewID, movieName, rating, path)
	renderEditReviewResult(w, movieName, rating, err)
}

// renderEditReviewResult renders the success/failure banner for an edit, along
// with the review list underneath it so the UI has somewhere to land either way.
func renderEditReviewResult(w http.ResponseWriter, movieName string, rating int, editErr error) {
	view := editReviewResultView{MovieName: movieName, Rating: rating, Err: editErr}

	revs, err := rr.FetchAllReviews()
	if err != nil {
		if view.Err == nil {
			view.Err = err
		}
	} else {
		view.Reviews = revs
	}

	if err := templates.ExecuteTemplate(w, "edit_review_result.html", view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serveAddReviewForm renders the initial "add a review" form.
func serveAddReviewForm(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if err := templates.ExecuteTemplate(w, "add_review_form.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// compareView is the data for one step of the comparison flow: the review being
// added, how far it's been walked into the tree so far, and what it's currently
// being compared against.
type compareView struct {
	MovieName string
	Rating    int
	Path      []string
	Current   *RatingTreeNode
}

// addReviewResultView is the data for the outcome of the comparison flow.
type addReviewResultView struct {
	MovieName string
	Rating    int
	Err       error
	Root      *RatingTreeNode
}

// compareReview handles one step of the add-review comparison flow: it's posted
// to both by the initial form (with an empty path) and by each comparison choice
// (with the path built up so far, plus this step's "left"/"right" pick appended
// via the clicked button's own name/value). Once the path reaches an empty slot,
// it performs the real insert and renders the result instead of another comparison.
func compareReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	movieName := r.Form.Get("movie_name")
	if len(movieName) == 0 {
		http.Error(w, "movie_name is required", http.StatusBadRequest)
		return
	}

	rating, err := strconv.Atoi(r.Form.Get("rating"))
	if err != nil {
		http.Error(w, "rating must be an integer", http.StatusBadRequest)
		return
	}

	path := r.Form["path"]

	node, err := NodeAtPath(&rr, rating, path)
	if err != nil {
		renderAddReviewResult(w, movieName, rating, err)
		return
	}

	if node != nil {
		view := compareView{MovieName: movieName, Rating: rating, Path: path, Current: node}
		if err := templates.ExecuteTemplate(w, "compare.html", view); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	review := &MovieReview{MovieName: movieName, Rating: rating}
	err = InsertReviewAtPath(&rr, review, path)
	renderAddReviewResult(w, movieName, rating, err)
}

// renderAddReviewResult renders the success/failure banner. On success it also
// re-fetches the rating's tree so the result includes the movie just added.
func renderAddReviewResult(w http.ResponseWriter, movieName string, rating int, insertErr error) {
	view := addReviewResultView{MovieName: movieName, Rating: rating, Err: insertErr}

	if insertErr == nil {
		reviews, err := rr.FetchReviewsByRating(rating)
		if err != nil {
			view.Err = err
		} else {
			view.Root = BuildTree(reviews)
		}
	}

	if err := templates.ExecuteTemplate(w, "add_review_result.html", view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const (
	minRating = 1
	maxRating = 10
)

// ratingBarView is one bar of the rating breakdown chart.
type ratingBarView struct {
	Rating    int
	Count     int
	HeightPct int // 0-100, relative to the highest count among all ratings
}

// ratingBreakdownView is the full chart: one bar per rating 1-10, plus whether
// there's any data at all (so the template can show an empty state instead of
// ten flat zero-height bars).
type ratingBreakdownView struct {
	Bars    []ratingBarView
	HasData bool
}

// serveRatingBreakdown renders the bar chart of how many films are rated at each
// score from 1 to 10.
func serveRatingBreakdown(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	counts, err := rr.CountReviewsByRating()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	max := 0
	for _, c := range counts {
		if c > max {
			max = c
		}
	}

	bars := make([]ratingBarView, 0, maxRating-minRating+1)
	for rating := minRating; rating <= maxRating; rating++ {
		count := counts[rating]
		heightPct := 0
		if max > 0 {
			heightPct = count * 100 / max
		}
		bars = append(bars, ratingBarView{Rating: rating, Count: count, HeightPct: heightPct})
	}

	view := ratingBreakdownView{Bars: bars, HasData: max > 0}

	if err := templates.ExecuteTemplate(w, "rating_breakdown.html", view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ratingRankingView is the ranked (best-to-worst) list of films at one rating.
type ratingRankingView struct {
	Rating  int
	Reviews []MovieReview
}

// serveRatingRanking renders the full ranking of films at one rating, highest to
// lowest, for when a bar in the breakdown chart is clicked.
func serveRatingRanking(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	rating, err := strconv.Atoi(r.URL.Query().Get("rating"))
	if err != nil {
		http.Error(w, "rating must be an integer", http.StatusBadRequest)
		return
	}

	reviews, err := rr.FetchReviewsByRating(rating)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	view := ratingRankingView{Rating: rating, Reviews: InOrderDescending(BuildTree(reviews))}

	if err := templates.ExecuteTemplate(w, "rating_ranking.html", view); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
