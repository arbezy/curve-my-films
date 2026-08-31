package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var rr = ReviewRepository{}

func HandleRequests() {
	http.Handle("/reviews", http.HandlerFunc(getAllReviews))
	http.Handle("/review", http.HandlerFunc(handleReview))
	http.Handle("/tree", http.HandlerFunc(getRatingTree))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		getReview(w, r)
	case http.MethodPost:
		addReview(w, r)
	case http.MethodDelete:
		deleteReview(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllReviews(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	revs, err := rr.FetchAllReviews()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting all reviews: %v", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(revs)
}

func getReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	r.ParseForm()
	movieName := r.Form.Get("moviename")

	if len(movieName) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Printf("Looking for movie: %s \n", movieName)

	rev, err := rr.FetchMovieReview(movieName)
	if err != nil {
		fmt.Printf("Error %v", err)
		http.Error(w, fmt.Sprintf("Error fetching review: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rev)
}

// TODO: could find a way to get rid of unnecessary info here, as we only really need movie name from
// from the movie review right now. Think I would need to make a new struct for this with json -'s (blanks)'?
func getRatingTree(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	r.ParseForm()
	ratingStr := r.Form.Get("rating")

	if len(ratingStr) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("No rating supplied with request")
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Cannot convert rating to integer")
		return
	}

	reviews, err := rr.FetchReviewsByRating(rating)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error during FetchReviewsByRating: %v", err), http.StatusInternalServerError)
		return
	}

	if len(reviews) == 0 {
		http.Error(w, "Length of fetched reviews is zero", http.StatusInternalServerError)
		return
	}

	reviewTreeRoot := BuildTree(reviews)
	if reviewTreeRoot == nil {
		http.Error(w, "root node of tree is nil", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ToResponse(reviewTreeRoot)); err != nil {
		fmt.Println("JSON encode error:", err)
	}
}

func deleteReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	r.ParseForm()
	idStr := r.Form.Get("id")
	if len(idStr) == 0 {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "id must be an integer", http.StatusBadRequest)
		return
	}

	target, err := rr.FetchReviewByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching review: %v", err), http.StatusNotFound)
		return
	}

	reviews, err := rr.FetchReviewsByRating(target.Rating)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching rating tree: %v", err), http.StatusInternalServerError)
		return
	}

	root := BuildTree(reviews)
	node := findNode(root, id)
	if node == nil {
		http.Error(w, "review not found in its rating tree", http.StatusInternalServerError)
		return
	}

	for _, n := range DeleteNode(node) {
		if err := rr.UpdateReviewPointers(n.Review.ReviewID, ptrID(n.Left), ptrID(n.Right), ptrID(n.Parent)); err != nil {
			http.Error(w, fmt.Sprintf("error updating tree pointers: %v", err), http.StatusInternalServerError)
			return
		}
	}

	log.Println("tree after deletion:")
	PrintTree(root)

	if err := rr.DeleteReview(id); err != nil {
		http.Error(w, fmt.Sprintf("error deleting review: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func addReview(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	var req AddReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.MovieName) == 0 {
		http.Error(w, "movie_name is required", http.StatusBadRequest)
		return
	}

	newReview := &MovieReview{
		MovieName: req.MovieName,
		Rating:    req.Rating,
	}

	existing, err := rr.FetchReviewsByRating(req.Rating)
	if err != nil {
		http.Error(w, fmt.Sprintf("error fetching existing reviews: %v", err), http.StatusInternalServerError)
		return
	}

	// no reviews at this rating yet -> new review becomes the root
	if len(existing) == 0 {
		if _, err := rr.InsertReview(newReview); err != nil {
			http.Error(w, fmt.Sprintf("error inserting review: %v", err), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		return
	}

	root := BuildTree(existing)
	if root == nil {
		http.Error(w, "root node of tree is nil", http.StatusInternalServerError)
		return
	}

	// walk the path to the empty slot the client resolved the new review to
	var parent *RatingTreeNode
	var side string
	node := root
	for _, direction := range req.Path {
		if node == nil {
			http.Error(w, "path does not lead to an empty position in the tree", http.StatusBadRequest)
			return
		}
		parent = node
		side = direction
		switch direction {
		case "left":
			node = node.Left
		case "right":
			node = node.Right
		default:
			http.Error(w, fmt.Sprintf("invalid path direction: %s", direction), http.StatusBadRequest)
			return
		}
	}

	if parent == nil {
		http.Error(w, "path must not be empty when the rating tree is non-empty", http.StatusBadRequest)
		return
	}
	if node != nil {
		http.Error(w, "position at end of path is already occupied", http.StatusConflict)
		return
	}

	newReview.ParentPtr = sql.NullInt64{Int64: int64(parent.Review.ReviewID), Valid: true}

	newID, err := rr.InsertReview(newReview)
	if err != nil {
		http.Error(w, fmt.Sprintf("error inserting review: %v", err), http.StatusInternalServerError)
		return
	}

	if err := rr.UpdateReviewChildPtr(int64(parent.Review.ReviewID), side, newID); err != nil {
		http.Error(w, fmt.Sprintf("error linking review to parent: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
