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
	"github.com/akaritrading/prices/pkg/pricesclient"
)

var pricesBinanceClient = &pricesclient.Client{
	Host: flag.PricesHost(),
}

var redisHandle *redis.Handle
var globalLogger *log.Logger

func main() {

	globalLogger = log.New("platform", "")

	pricesBinanceClient.InitBinance()

	db := initDB()
	migrate(db)

	redisHandle = redis.DefaultConnect()
	defer redisHandle.Close()

	engineclient.Init(redisHandle)

	r := chi.NewRouter()
	r.Use(middleware.RequestContext("platform", db))
	r.Use(middleware.Recoverer)

	r.Route("/api/auth", AuthRoutes)
	r.Route("/api", apiRoute)
	r.Route("/ws", wsRoute)

	server := &http.Server{
		Addr:    flag.PlatformHost(),
		Handler: r,
	}

	globalLogger.Fatal(server.ListenAndServe())
}

func apiRoute(r chi.Router) {
	r.Use(jsonResponse)
	r.Use(authentication)
	r.Route("/scripts", ScriptRoute)
	r.Route("/history", HistoryRoute)
	r.Route("/scripts/{scriptID}/versions", ScriptVersionsRoute)
	r.Route("/jobs", JobsRoute)
	r.Route("/trades", TradesRoute)
	r.Route("/userExchanges", ExchangesRoute)

}

func wsRoute(r chi.Router) {
	r.Use(authentication)
	r.Get("/backtest", backtest)
}

func migrate(d *db.DB) error {
	return d.Gorm().AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.Job{},
		&db.PendingUser{},
		&db.Credential{},
		&db.ExchangeConnection{},
		&db.User{})
}

func initDB() *db.DB {
	db, err := db.DefaultOpen()
	if err != nil {
		globalLogger.Fatal(errors.New("platform could not connect to postgres. exiting."))
	}
	return db
}
