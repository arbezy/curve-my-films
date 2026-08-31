package main

import (
	"database/sql"
	"errors"
	"fmt"
)

var (
	ErrEmptyPath     = errors.New("path must not be empty when the rating tree is non-empty")
	ErrInvalidPath   = errors.New("path does not lead to an empty position in the tree")
	ErrPositionTaken = errors.New("position at end of path is already occupied")
)

// ptrID converts a node into the sql.NullInt64 form its review_id takes in a
// left_ptr/right_ptr/parent_ptr column (NULL for a nil node).
func ptrID(node *RatingTreeNode) sql.NullInt64 {
	if node == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(node.Review.ReviewID), Valid: true}
}

// InsertReviewAtPath finds the empty tree slot addressed by path (a sequence of
// "left"/"right" directions from the root) and inserts review there, persisting
// the new row and linking it to its parent. If the rating has no reviews yet,
// review becomes the root and path is ignored.
func InsertReviewAtPath(rr *ReviewRepository, review *MovieReview, path []string) error {
	existing, err := rr.FetchReviewsByRating(review.Rating)
	if err != nil {
		return err
	}

	if len(existing) == 0 {
		_, err := rr.InsertReview(review)
		return err
	}

	root := BuildTree(existing)

	var parent *RatingTreeNode
	var side string
	node := root
	for _, direction := range path {
		if node == nil {
			return ErrInvalidPath
		}
		parent, side = node, direction
		switch direction {
		case "left":
			node = node.Left
		case "right":
			node = node.Right
		default:
			return fmt.Errorf("invalid path direction: %s", direction)
		}
	}

	if parent == nil {
		return ErrEmptyPath
	}
	if node != nil {
		return ErrPositionTaken
	}

	review.ParentPtr = sql.NullInt64{Int64: int64(parent.Review.ReviewID), Valid: true}

	newID, err := rr.InsertReview(review)
	if err != nil {
		return err
	}
	return rr.UpdateReviewChildPtr(int64(parent.Review.ReviewID), side, newID)
}

// DeleteReviewByID removes the review with the given ID from its rating tree,
// rewiring the tree around it (see DeleteNode), and deletes its row.
func DeleteReviewByID(rr *ReviewRepository, reviewID int) error {
	target, err := rr.FetchReviewByID(reviewID)
	if err != nil {
		return err
	}

	reviews, err := rr.FetchReviewsByRating(target.Rating)
	if err != nil {
		return err
	}

	root := BuildTree(reviews)
	node := findNode(root, reviewID)
	if node == nil {
		return fmt.Errorf("review %d not found in its rating tree", reviewID)
	}

	for _, n := range DeleteNode(node) {
		if err := rr.UpdateReviewPointers(n.Review.ReviewID, ptrID(n.Left), ptrID(n.Right), ptrID(n.Parent)); err != nil {
			return err
		}
	}

	return rr.DeleteReview(reviewID)
}
