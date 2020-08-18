package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jinzhu/gorm"

	"github.com/akaritrading/libs/db"
)

var DB *gorm.DB

func main() {

	db, err := db.Open("localhost", "postgres", "postgres", "password")
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer db.Close()

	DB.LogMode(true)

	if err = migrate(); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)     // replace eventually
	r.Use(middleware.DefaultLogger) // replace eventually

	r.Route("/api", apiRoute)

	server := &http.Server{
		Addr:    ":6060",
		Handler: r,
	}
	log.Fatal(server.ListenAndServe())
}

func apiRoute(r chi.Router) {

	r.Use(authentication) // authentication middleware
	r.Use(jsonResponse)   // adds json content header
	r.Route("/scripts", ScriptRoute)
	r.Route("/scripts/{id}/versions", ScriptVersionsRoute)
}

func migrate() error {
	// creates tables and new columns for existing tables
	return DB.AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.User{}).Error
}
