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
	leftPtr   int    `json:"left_ptr"`
	rightPtr  int    `json:"right_ptr"`
}

type RatingTreeNode struct {
	Review MovieReview     `json:"value"`
	Left   *RatingTreeNode `json:"left"`
	Right  *RatingTreeNode `json:"right"`
}

var rr = ReviewRepository{}

func getReview(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	movieName := r.Form.Get("moviename")

	if len(movieName) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rev, err := rr.ReadMovieReview(movieName)
	if err != nil {
		log.Fatal(err)
	}

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
	var parentRev = MovieReview{ReviewID: 0, MovieName: "testmovie1", Rating: rating}
	var childRev0 = MovieReview{ReviewID: 1, MovieName: "testmovie2", Rating: rating}
	var childRev1 = MovieReview{ReviewID: 2, MovieName: "testmovie3", Rating: rating}

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
	reviewRepo, err := InitDB()
	rr = reviewRepo
	if err != nil {
		log.Fatal(err)
	}
	handleRequests()
}
