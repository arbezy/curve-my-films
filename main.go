package main

// TODO: put in cmd dir
// TODO: move all handlers into a handlers.go file !

import (
	"encoding/json"
	"github.com/arbezy/curve-my-films/config"
	"log"
	"net/http"
	"strconv"
)

var rr = ReviewRepository{}

func getReview(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	movieName := r.Form.Get("moviename")

	if len(movieName) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rev, err := rr.FetchMovieReview(movieName)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rev)
}

// TODO: could find a way to get rid of unnecessary info here, as we only really need movie name from
// from the movie review right now. Think I would need to make a new struct for this with json -'s (blanks)'?
func getRatingTree(w http.ResponseWriter, r *http.Request) {
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

	// TODO: replace with db fetch -> this has been implemented now as FetchReviewsByRating
	var parentRev = MovieReview{ReviewID: 0, MovieName: "testmovie1", Rating: rating}
	var childRev0 = MovieReview{ReviewID: 1, MovieName: "testmovie2", Rating: rating}
	var childRev1 = MovieReview{ReviewID: 2, MovieName: "testmovie3", Rating: rating}

	// contruct the tree from the fetched reviews
	var node0 = RatingTreeNode{Review: childRev0, Left: nil, Right: nil}
	var node1 = RatingTreeNode{Review: childRev1, Left: nil, Right: nil}
	root := RatingTreeNode{Review: parentRev, Left: &node0, Right: &node1}
	node0.Parent = &root
	node1.Parent = &root

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(root)
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
	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	reviewRepo, err := InitDB()
	rr = reviewRepo
	if err != nil {
		log.Fatal(err)
	}

	handleRequests()
}
