package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/redis"
)

var DB *db.DB
var redisHandle *redis.Handle
var DebugEnginePort = ":7979" // remove in production

func main() {

	DB = initDB()
	migrate()
	defer DB.Gorm().Close()

	redisHandle = initRedis()
	redisHandle.Connect()
	defer redisHandle.Close()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)     // replace eventually
	r.Use(middleware.DefaultLogger) // replace eventually
	r.Use(middleware.RequestID)     //

	r.Route("/api", apiRoute)
	r.Route("/ws", wsRoute)

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
	return DB.Gorm().AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.User{}).Error
}

func initDB() *db.DB {
	db, err := db.Open("localhost", "postgres", "postgres", "password")
	if err != nil {
		log.Fatal(err)
	}
	db.Gorm().LogMode(true)
	return db
}

func initRedis() *redis.Handle {
	return &redis.Handle{
		MaxActive:       100,
		MaxIdle:         30,
		IdleTimeout:     time.Minute * 5,
		MaxConnLifetime: time.Minute * 5,
	}
}
