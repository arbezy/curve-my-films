# Curve My Films
Work in progress movie review app.

The goal is to create a reviewing app that will prompt a user to rank films within a rating, so you can get a more _pure_ rating. I.e. if rating two 8/10 films user would be prompted to rank them and so get a "high" 8 and a "low" 8 rating for those two films. I do this by creating a BST for each rating, then when adding a film we walk this tree doing comparisons to preexisiting ratings to find where a new movie should sit.

Aiming to use just the Go stdlib and htmx if I can (might have to add some JS if unavoidable), as I think it is more interesting and a better learning opportunity. Who even needs a framework in go anyway.

## Setup

### Prerequisites
- Go 1.25.5 or later
- A running MySQL server

### 1. Configure the database connection
Copy `.env.example` to `.env` and fill in your database credentials:
```
cp .env.example .env
```

### 2. Create the database and table
```sql
CREATE DATABASE curvemyfilms;
USE curvemyfilms;

CREATE TABLE reviews (
  review_id  INT NOT NULL AUTO_INCREMENT,
  movie_name VARCHAR(50) NOT NULL,
  rating     INT NOT NULL,
  left_ptr   INT DEFAULT NULL,
  right_ptr  INT DEFAULT NULL,
  parent_ptr INT DEFAULT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (review_id)
);
```

### 3. (optional) Install air
```brew install air```

## Running
To run use:
`go run .`
in root dir

or `air` to run with hot reload! (if you have it installed)
