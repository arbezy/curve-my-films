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
	ReviewID  int         `json:"review_id"`
	ParentId  OptionalInt `json:"parent_id"`
	MovieName string      `json:"movie_name"`
	Rating    int         `json:"rating"`
}

func getReview(w http.ResponseWriter, r *http.Request) {
	var rev = MovieReview{0, OptionalInt{}, "testmovie", 5}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rev)
}

func handleRequests() {
	http.Handle("/review", http.HandlerFunc(getReview))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
