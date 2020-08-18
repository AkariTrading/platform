package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jinzhu/gorm"

	"github.com/akaritrading/libs/db"
)

// Handle that holds connections to Redis, Postrgres, used globally
var DB *gorm.DB

func main() {

	db, err := db.Open("localhost", "postgres", "postgres", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.DefaultLogger)

	r.Route("/api", apiRoute)

	server := &http.Server{
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())
}

func apiRoute(r chi.Router) {

	r.Use() // add authentication middleware
	r.Route("/scripts", ScriptRoute)
	r.Route("/scriptVersions", ScriptVersionsRoute)
}
