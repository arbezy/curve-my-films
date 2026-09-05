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

// walkPath follows path (a sequence of "left"/"right" directions) from root and
// returns the node it lands on (nil if it reaches an empty slot), along with the
// last node stepped through (parent) and which side of it was taken (side) — nil
// and "" if path is empty. Returns ErrInvalidPath if the path runs past a nil node.
func walkPath(root *RatingTreeNode, path []string) (parent *RatingTreeNode, side string, node *RatingTreeNode, err error) {
	node = root
	for _, direction := range path {
		if node == nil {
			return nil, "", nil, ErrInvalidPath
		}
		parent, side = node, direction
		switch direction {
		case "left":
			node = node.Left
		case "right":
			node = node.Right
		default:
			return nil, "", nil, fmt.Errorf("invalid path direction: %s", direction)
		}
	}
	return parent, side, node, nil
}

// NodeAtPath returns the node currently at path within rating's tree, e.g. so the
// frontend can show which review the new one is being compared against. A nil
// node with a nil error means path has reached an empty slot (or the rating has
// no reviews at all) — i.e. it's ready to be inserted there.
func NodeAtPath(rr *ReviewRepository, rating int, path []string) (*RatingTreeNode, error) {
	existing, err := rr.FetchReviewsByRating(rating)
	if err != nil {
		return nil, err
	}
	if len(existing) == 0 {
		return nil, nil
	}

	root := BuildTree(existing)
	_, _, node, err := walkPath(root, path)
	return node, err
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

	parent, side, node, err := walkPath(root, path)
	if err != nil {
		return err
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

// detachFromTree removes reviewID's node from its current rating tree, rewiring
// Left/Right/Parent pointers on the surrounding nodes to preserve BST structure
// (see DeleteNode). reviewID's own row is left untouched - its pointers are
// stale once detached, so callers must either delete the row or overwrite its
// pointers before it's read again. Returns the review as it stood (with its
// pre-detach rating) for callers that still need that.
func detachFromTree(rr *ReviewRepository, reviewID int) (*MovieReview, error) {
	target, err := rr.FetchReviewByID(reviewID)
	if err != nil {
		return nil, err
	}

	reviews, err := rr.FetchReviewsByRating(target.Rating)
	if err != nil {
		return nil, err
	}

	root := BuildTree(reviews)
	node := findNode(root, reviewID)
	if node == nil {
		return nil, fmt.Errorf("review %d not found in its rating tree", reviewID)
	}

	for _, n := range DeleteNode(node) {
		if err := rr.UpdateReviewPointers(n.Review.ReviewID, ptrID(n.Left), ptrID(n.Right), ptrID(n.Parent)); err != nil {
			return nil, err
		}
	}

	return target, nil
}

// DeleteReviewByID removes the review with the given ID from its rating tree,
// rewiring the tree around it, and deletes its row.
func DeleteReviewByID(rr *ReviewRepository, reviewID int) error {
	if _, err := detachFromTree(rr, reviewID); err != nil {
		return err
	}
	return rr.DeleteReview(reviewID)
}

// RepositionReviewAtPath applies an edit that changes reviewID's rating: it
// detaches the review from its current rating tree, updates its movie name and
// rating, and re-inserts it into the new rating's tree at the empty slot
// addressed by path — the same walk used when adding a brand new review, so an
// edited review gets re-ranked via the same comparison flow rather than landing
// at an arbitrary position.
func RepositionReviewAtPath(rr *ReviewRepository, reviewID int, movieName string, rating int, path []string) error {
	if _, err := detachFromTree(rr, reviewID); err != nil {
		return err
	}

	if err := rr.UpdateReviewDetails(reviewID, movieName, rating); err != nil {
		return err
	}

	all, err := rr.FetchReviewsByRating(rating)
	if err != nil {
		return err
	}
	existing := make([]*MovieReview, 0, len(all))
	for _, rev := range all {
		if rev.ReviewID != reviewID {
			existing = append(existing, rev)
		}
	}

	if len(existing) == 0 {
		return rr.UpdateReviewPointers(reviewID, sql.NullInt64{}, sql.NullInt64{}, sql.NullInt64{})
	}

	root := BuildTree(existing)

	parent, side, node, err := walkPath(root, path)
	if err != nil {
		return err
	}
	if parent == nil {
		return ErrEmptyPath
	}
	if node != nil {
		return ErrPositionTaken
	}

	parentPtr := sql.NullInt64{Int64: int64(parent.Review.ReviewID), Valid: true}
	if err := rr.UpdateReviewPointers(reviewID, sql.NullInt64{}, sql.NullInt64{}, parentPtr); err != nil {
		return err
	}
	return rr.UpdateReviewChildPtr(int64(parent.Review.ReviewID), side, int64(reviewID))
}
