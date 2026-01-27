package main

// TODO: put in cmd dir

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type MovieReview struct {
	ReviewID  int    `json:"review_id"`
	MovieName string `json:"movie_name"`
	Rating    int    `json:"rating"`
}

type RatingTreeNode struct {
	Review MovieReview     `json:"value"`
	Left   *RatingTreeNode `json:"left"`
	Right  *RatingTreeNode `json:"right"`
}

func getReview(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	movieName := r.Form.Get("moviename")

	if len(movieName) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: replace with db fetch
	var rev = MovieReview{0, movieName, 10}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rev)
}

func getRatingTree(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ratingStr := r.Form.Get("rating")

	if len(ratingStr) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rating, err := strconv.Atoi(ratingStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Cannot convert rating to integer")
		return
	}

	// TODO: replace with db fetch
	var parentRev = MovieReview{0, "testmovie1", rating}
	var childRev0 = MovieReview{1, "testmovie2", rating}
	var childRev1 = MovieReview{2, "testmovie3", rating}

	var node0 = RatingTreeNode{childRev0, nil, nil}
	var node1 = RatingTreeNode{childRev1, nil, nil}

	tree := RatingTreeNode{Review: parentRev, Left: &node0, Right: &node1}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

// TODO: supplied with an updated tree ?
func updateRatingTree(w http.ResponseWriter, r *http.Request) {
	// TODO: parse new tree from request body
	// TODO: perform db update

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success") // NOTE: if new tree not created in the front end then return new tree here
}

func handleRequests() {
	http.Handle("/review", http.HandlerFunc(getReview))
	http.Handle("/tree", http.HandlerFunc(getRatingTree))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
