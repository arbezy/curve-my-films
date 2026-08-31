package main

import (
	"encoding/json"
	"errors"
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

	switch err := DeleteReviewByID(&rr, id); {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrReviewNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, fmt.Sprintf("error deleting review: %v", err), http.StatusInternalServerError)
	}
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

	review := &MovieReview{
		MovieName: req.MovieName,
		Rating:    req.Rating,
	}

	switch err := InsertReviewAtPath(&rr, review, req.Path); {
	case err == nil:
		w.WriteHeader(http.StatusCreated)
	case errors.Is(err, ErrPositionTaken):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, ErrInvalidPath), errors.Is(err, ErrEmptyPath):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, fmt.Sprintf("error inserting review: %v", err), http.StatusInternalServerError)
	}
}
