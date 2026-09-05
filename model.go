package main

import (
	"database/sql"
	"time"
)

type MovieReview struct {
	ReviewID  int           `json:"review_id"`
	MovieName string        `json:"movie_name"`
	Rating    int           `json:"rating"`
	LeftPtr   sql.NullInt64 `json:"left_ptr"`
	RightPtr  sql.NullInt64 `json:"right_ptr"`
	ParentPtr sql.NullInt64 `json:"parent_ptr"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type RatingTreeNode struct {
	Review MovieReview     `json:"value"`
	Left   *RatingTreeNode `json:"left"`
	Right  *RatingTreeNode `json:"right"`
	Parent *RatingTreeNode `json:"-"` // tag prevents cycles
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
