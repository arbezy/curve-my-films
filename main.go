package main

// TODO: put in cmd dir
// TODO: move all handlers into a handlers.go file !

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/arbezy/curve-my-films/config"
)

var rr = ReviewRepository{}

func getAllReviews(w http.ResponseWriter, r *http.Request) {
	revs, err := rr.FetchAllReviews()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting all reviews: %v", err), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(revs)
}

func getReview(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, fmt.Sprintf("Length of fetched reviews is zero"), http.StatusInternalServerError)
		return
	}

	reviewTreeRoot := buildTree(reviews)
	if reviewTreeRoot == nil {
		http.Error(w, fmt.Sprint("root node of tree is nil"), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ToResponse(reviewTreeRoot)); err != nil {
		fmt.Println("JSON encode error:", err)
	}
}

func buildTree(reviews []*MovieReview) *RatingTreeNode {
	fmt.Println("building tree..")
	if len(reviews) == 0 {
		return nil
	}

	// create a node for each review, mapped by ReviewID
	nodeMap := make(map[int]*RatingTreeNode)
	for _, review := range reviews {
		nodeMap[review.ReviewID] = &RatingTreeNode{
			Review: *review,
		}
	}

	// wire up the pointers
	var root *RatingTreeNode
	for _, review := range reviews {
		node := nodeMap[review.ReviewID]

		if review.LeftPtr.Valid {
			node.Left = nodeMap[int(review.LeftPtr.Int64)]
		}
		if review.RightPtr.Valid {
			node.Right = nodeMap[int(review.RightPtr.Int64)]
		}
		if review.ParentPtr.Valid {
			node.Parent = nodeMap[int(review.ParentPtr.Int64)]
		} else {
			// no parent means this is the root
			root = node
		}
	}

	return root
}

// TODO: supplied with an updated tree ?
// NOTE: actually I'd rather do just an add rating function...
func updateRatingTree(w http.ResponseWriter, r *http.Request) {
	// TODO: parse new tree from request body
	// TODO: perform db update

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("success") // NOTE: if new tree not created in the front end then return new tree here
}

// TODO: add rating (to tree)
func addReview(w http.ResponseWriter, r *http.Request) {
	// parse review from request
	// find relevent rating tree
	// add to tree -> but how? since we don't know where it should go
	// also pass in request , where the new review should go in tree. This could be done with a list of lefts and rights?
	// would then need to rebalance tree I think
	// then need to write this back to the db, the easiest way to do this would be to delete the entire tree and replace with new tree
	// there is probably a more clever way to do this but I would need to calculate the actions to take on the existing tree to update
	// to match new rating structure.
}

func handleRequests() {
	http.Handle("/reviews", http.HandlerFunc(getAllReviews))
	http.Handle("/review", http.HandlerFunc(handleReview))
	http.Handle("/tree", http.HandlerFunc(getRatingTree))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleReview(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getReview(w, r)
	case http.MethodPost:
		addReview(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	fmt.Println("Running CurveMyFilms")

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
