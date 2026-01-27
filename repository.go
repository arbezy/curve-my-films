package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
)

type ReviewRepository struct {
	db *sql.DB
}

// TODO: move this into a config.go file and read these values from a .env file
func GetConfig() *mysql.Config {
	var cfg = mysql.NewConfig()
	cfg.User = 
	cfg.Passwd = 
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "curvemyfilms"

	return cfg
}

// TODO: LEFT OFF HERE! wire this up to the handler
func (rr *ReviewRepository) ReadMovieReview(movieName string) (*MovieReview, error) {
	query := fmt.Sprintf("SELECT * FROM reviews WHERE movie_name=%s;", movieName)
	results, err := rr.db.Query(query)
	if err != nil {
		return nil, err
	}

	rev := &MovieReview{}
	if results.Next() {
		err = results.Scan(&rev.ReviewID, &rev.MovieName, &rev.Rating)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("review not found")
	}

	return rev, nil
}
