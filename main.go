package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type OptionalInt struct {
	val int
}

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
	var rev = MovieReview{0, "testmovie", 5}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rev)
}

func getRatingTree(w http.ResponseWriter, r *http.Request) {
	var parentRev = MovieReview{0, "testmovie1", 5}
	var childRev0 = MovieReview{1, "testmovie2", 5}
	var childRev1 = MovieReview{2, "testmovie3", 5}

	var node0 = RatingTreeNode{childRev0, nil, nil}
	var node1 = RatingTreeNode{childRev1, nil, nil}

	tree := RatingTreeNode{Review: parentRev, Left: &node0, Right: &node1}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

func handleRequests() {
	http.Handle("/review", http.HandlerFunc(getReview))
	http.Handle("/tree", http.HandlerFunc(getRatingTree))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
