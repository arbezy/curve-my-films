package main

import "database/sql"

type MovieReview struct {
	ReviewID  int           `json:"review_id"`
	MovieName string        `json:"movie_name"`
	Rating    int           `json:"rating"`
	LeftPtr   sql.NullInt64 `json:"left_ptr"`
	RightPtr  sql.NullInt64 `json:"right_ptr"`
	ParentPtr sql.NullInt64 `json:"parent_ptr"`
}

type RatingTreeNode struct {
	Review MovieReview     `json:"value"`
	Left   *RatingTreeNode `json:"left"`
	Right  *RatingTreeNode `json:"right"`
	Parent *RatingTreeNode `json:"-"`
}

type RatingTreeNodeResponse struct {
	Review MovieReview             `json:"value"`
	Left   *RatingTreeNodeResponse `json:"left"`
	Right  *RatingTreeNodeResponse `json:"right"`
}

func ToResponse(node *RatingTreeNode) *RatingTreeNodeResponse {
	if node == nil {
		return nil
	}
	return &RatingTreeNodeResponse{
		Review: node.Review,
		Left:   ToResponse(node.Left),
		Right:  ToResponse(node.Right),
	}
}

type AddReviewRequest struct {
	MovieName string   `json:"movie_name"`
	Rating    int      `json:"rating"`
	Path      []string `json:"path"` // "left"/"right" directions from the root to the new review's slot
}

type UpdateReviewRequest struct {
	ReviewID  int    `json:"review_id"`
	MovieName string `json:"movie_name"`
}
