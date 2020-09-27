package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/flag"
	"github.com/akaritrading/libs/log"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/redis"
)

var redisHandle *redis.Handle

func main() {

	log.Init()

	db := initDB()
	migrate(db)

	redisHandle = redis.DefaultConnect()
	defer redisHandle.Close()

	engineclient.Init(redisHandle)

	r := chi.NewRouter()
	r.Use(middleware.RequestContext("platform", db))
	r.Use(middleware.Recoverer)

	r.Route("/auth", AuthRoutes)
	r.Route("/api", apiRoute)
	r.Route("/ws", wsRoute)

	server := &http.Server{
		Addr:    flag.PlatformHost(),
		Handler: r,
	}

	server.ListenAndServe()
}

func apiRoute(r chi.Router) {
	r.Use(jsonResponse)
	r.Use(authentication)
	r.Route("/scripts", ScriptRoute)
	r.Route("/history", HistoryRoute)
	r.Route("/scripts/{scriptID}/versions", ScriptVersionsRoute)
	r.Route("/jobs", JobsRoute)
}

func wsRoute(r chi.Router) {
	r.Use(authentication)
	r.Get("/backtest", backtest)
}

func migrate(d *db.DB) error {
	return d.Gorm().AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.ScriptJob{},
		&db.PendingUser{},
		&db.Credential{},
		&db.User{})
}

func initDB() *db.DB {
	db, err := db.DefaultOpen()
	if err != nil {
		log.Default().Fatal(errors.New("platform could not connect to postgres. exiting."))
	}
	return db
}
