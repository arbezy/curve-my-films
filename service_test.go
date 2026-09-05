package main

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

var reviewColumns = []string{"review_id", "movie_name", "rating", "left_ptr", "right_ptr", "parent_ptr", "created_at", "updated_at"}

var testTimestamp = time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

func newMockRepo(t *testing.T) (*ReviewRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &ReviewRepository{db: db}, mock
}

func TestInsertReviewAtPath_EmptyRating_BecomesRoot(t *testing.T) {
	rr, mock := newMockRepo(t)

	mock.ExpectQuery("SELECT * FROM reviews WHERE rating=?;").
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows(reviewColumns))

	mock.ExpectExec("INSERT INTO reviews (movie_name, rating, left_ptr, right_ptr, parent_ptr) VALUES (?, ?, ?, ?, ?);").
		WillReturnResult(sqlmock.NewResult(1, 1))

	review := &MovieReview{MovieName: "Aliens", Rating: 8}
	if err := InsertReviewAtPath(rr, review, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestInsertReviewAtPath_ExistingTree_InsertsAtPath(t *testing.T) {
	rr, mock := newMockRepo(t)

	mock.ExpectQuery("SELECT * FROM reviews WHERE rating=?;").
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows(reviewColumns).
			AddRow(1, "Alien", 8, nil, nil, nil, testTimestamp, testTimestamp))

	mock.ExpectExec("INSERT INTO reviews (movie_name, rating, left_ptr, right_ptr, parent_ptr) VALUES (?, ?, ?, ?, ?);").
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectExec("UPDATE reviews SET right_ptr = ? WHERE review_id = ?;").
		WithArgs(2, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	review := &MovieReview{MovieName: "Aliens", Rating: 8}
	if err := InsertReviewAtPath(rr, review, []string{"right"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteReviewByID_Leaf(t *testing.T) {
	rr, mock := newMockRepo(t)

	mock.ExpectQuery("SELECT * FROM reviews WHERE review_id=?;").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows(reviewColumns).
			AddRow(2, "Alien 3", 8, nil, nil, 1, testTimestamp, testTimestamp))

	mock.ExpectQuery("SELECT * FROM reviews WHERE rating=?;").
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows(reviewColumns).
			AddRow(1, "Aliens", 8, 2, nil, nil, testTimestamp, testTimestamp).
			AddRow(2, "Alien 3", 8, nil, nil, 1, testTimestamp, testTimestamp))

	mock.ExpectExec("UPDATE reviews SET left_ptr = ?, right_ptr = ?, parent_ptr = ? WHERE review_id = ?;").
		WithArgs(nil, nil, nil, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("DELETE FROM reviews WHERE review_id = ?;").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := DeleteReviewByID(rr, 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
