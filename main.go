package main

// TODO: put in cmd dir
// TODO: move all handlers into a handlers.go file !

import (
	"fmt"
	"log"

	"github.com/arbezy/curve-my-films/config"
)

func main() {
	fmt.Println("Running CurveMyFilms")

	err := config.Init()
	if err != nil {
		log.Fatal(err)
	}

	reviewRepo, err := InitDB()
	rr = reviewRepo
	if err != nil {
		log.Fatal(err)
	}

	HandleRequests()
}
