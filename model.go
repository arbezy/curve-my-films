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
