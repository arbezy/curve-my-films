package main

import (
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
