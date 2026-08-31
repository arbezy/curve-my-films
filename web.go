package main

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates = template.Must(template.ParseGlob(filepath.Join("templates", "*.html")))

// serveIndex serves the HTMX-driven frontend's single HTML page.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serveReviewsFragment renders the review list as an HTML fragment for HTMX to swap in.
func serveReviewsFragment(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s", r.Method, r.URL.Path)

	revs, err := rr.FetchAllReviews()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templates.ExecuteTemplate(w, "reviews.html", revs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
