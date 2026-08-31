package main

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/arbezy/curve-my-films/config"
	"github.com/go-sql-driver/mysql"
)

type ReviewRepository struct {
	db *sql.DB
}

// TODO: should probs move this to a different file too (db.go ?)
func GetDBConfig() *mysql.Config {
	var cfg = mysql.NewConfig()
	cfg.User = config.DatabaseConfig.DB_USERNAME
	cfg.Passwd = config.DatabaseConfig.DB_PASSWORD
	cfg.Net = config.DatabaseConfig.DB_NET
	cfg.Addr = config.DatabaseConfig.DB_ADDRESS
	cfg.DBName = config.DatabaseConfig.DB_NAME

	return cfg
}

// TODO: should probs move this to a different file too (db.go ?)
func InitDB() (ReviewRepository, error) {
	db, err := sql.Open("mysql", GetDBConfig().FormatDSN())
	if err != nil {
		return ReviewRepository{}, err
	}
	return ReviewRepository{db}, nil
}

func (rr *ReviewRepository) FetchAllReviews() ([]*MovieReview, error) {
	query := "SELECT * FROM reviews;"
	results, err := rr.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	reviews := []*MovieReview{}
	for results.Next() {
		rev := &MovieReview{}
		err = results.Scan(&rev.ReviewID, &rev.MovieName, &rev.Rating, &rev.LeftPtr, &rev.RightPtr, &rev.ParentPtr)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (rr *ReviewRepository) FetchMovieReview(movieName string) (*MovieReview, error) {
	query := "SELECT * FROM reviews WHERE movie_name=?;"
	results, err := rr.db.Query(query, movieName)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	rev := &MovieReview{}
	if results.Next() {
		err = results.Scan(&rev.ReviewID, &rev.MovieName, &rev.Rating, &rev.LeftPtr, &rev.RightPtr, &rev.ParentPtr)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("review not found")
	}

	return rev, nil
}

// inserts a new review row and returns its generated review_id.
func (rr *ReviewRepository) InsertReview(review *MovieReview) (int64, error) {
	query := "INSERT INTO reviews (movie_name, rating, left_ptr, right_ptr, parent_ptr) VALUES (?, ?, ?, ?, ?);"
	result, err := rr.db.Exec(query, review.MovieName, review.Rating, review.LeftPtr, review.RightPtr, review.ParentPtr)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// links a review to its parent by writing childID into the parent's left_ptr or right_ptr column.
func (rr *ReviewRepository) UpdateReviewChildPtr(parentID int64, side string, childID int64) error {
	var query string
	switch side {
	case "left":
		query = "UPDATE reviews SET left_ptr = ? WHERE review_id = ?;"
	case "right":
		query = "UPDATE reviews SET right_ptr = ? WHERE review_id = ?;"
	default:
		return fmt.Errorf("invalid side: %s", side)
	}
	_, err := rr.db.Exec(query, childID, parentID)
	return err
}

// reads all the reviews by rating so they can be contructed into a big TREE !
func (rr *ReviewRepository) FetchReviewsByRating(rating int) ([]*MovieReview, error) {
	query := "SELECT * FROM reviews WHERE rating=?;"
	results, err := rr.db.Query(query, rating)
	if err != nil {
		return nil, err
	}

	reviews := []*MovieReview{}
	for results.Next() {
		rev := &MovieReview{}
		err = results.Scan(&rev.ReviewID, &rev.MovieName, &rev.Rating, &rev.LeftPtr, &rev.RightPtr, &rev.ParentPtr)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, rev)
	}

	return reviews, nil
}
