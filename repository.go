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

var ErrReviewNotFound = errors.New("review not found")

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
		return nil, ErrReviewNotFound
	}

	return rev, nil
}

// fetches a single review by its review_id.
func (rr *ReviewRepository) FetchReviewByID(reviewID int) (*MovieReview, error) {
	query := "SELECT * FROM reviews WHERE review_id=?;"
	row := rr.db.QueryRow(query, reviewID)

	rev := &MovieReview{}
	err := row.Scan(&rev.ReviewID, &rev.MovieName, &rev.Rating, &rev.LeftPtr, &rev.RightPtr, &rev.ParentPtr)
	if err == sql.ErrNoRows {
		return nil, ErrReviewNotFound
	}
	if err != nil {
		return nil, err
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

// overwrites a review's left/right/parent pointers, e.g. after rewiring the tree around a deletion.
func (rr *ReviewRepository) UpdateReviewPointers(reviewID int, left, right, parent sql.NullInt64) error {
	query := "UPDATE reviews SET left_ptr = ?, right_ptr = ?, parent_ptr = ? WHERE review_id = ?;"
	_, err := rr.db.Exec(query, left, right, parent, reviewID)
	return err
}

// updates a review's movie_name. Returns ErrReviewNotFound if no row matches reviewID.
func (rr *ReviewRepository) UpdateReviewName(reviewID int, movieName string) error {
	query := "UPDATE reviews SET movie_name = ? WHERE review_id = ?;"
	result, err := rr.db.Exec(query, movieName, reviewID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrReviewNotFound
	}
	return nil
}

// deletes a single review row.
func (rr *ReviewRepository) DeleteReview(reviewID int) error {
	query := "DELETE FROM reviews WHERE review_id = ?;"
	_, err := rr.db.Exec(query, reviewID)
	return err
}

// counts reviews per rating, for the rating breakdown chart. Ratings with no
// reviews are simply absent from the returned map.
func (rr *ReviewRepository) CountReviewsByRating() (map[int]int, error) {
	query := "SELECT rating, COUNT(*) FROM reviews GROUP BY rating;"
	results, err := rr.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer results.Close()

	counts := make(map[int]int)
	for results.Next() {
		var rating, count int
		if err := results.Scan(&rating, &count); err != nil {
			return nil, err
		}
		counts[rating] = count
	}

	return counts, nil
}

// reads all the reviews by rating so they can be contructed into a big TREE !
func (rr *ReviewRepository) FetchReviewsByRating(rating int) ([]*MovieReview, error) {
	query := "SELECT * FROM reviews WHERE rating=?;"
	results, err := rr.db.Query(query, rating)
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
