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
	query := fmt.Sprintf("SELECT * FROM reviews;")
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
	query := fmt.Sprintf("SELECT * FROM reviews WHERE movie_name='%s';", movieName)
	results, err := rr.db.Query(query)
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

// Reads all the reviews by rating so they can be contructed into a big TREE !
func (rr *ReviewRepository) FetchReviewsByRating(rating int) ([]*MovieReview, error) {
	query := fmt.Sprintf("SELECT * FROM reviews WHERE rating=%d;", rating)
	results, err := rr.db.Query(query)
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
