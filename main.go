package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/log"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/redis"
	"github.com/akaritrading/libs/util"
)

var DB *db.DB
var redisHandle *redis.Handle
var port = ":6060"

func main() {

	log.Init()

	DB = initDB()
	migrate()

	redisHandle = initRedis()
	redisHandle.Connect()
	defer redisHandle.Close()

	engineclient.Init(redisHandle)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger("platform"))
	r.Use(middleware.Recoverer)

	r.Route("/api", apiRoute)
	r.Route("/ws", wsRoute)

	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	println("Now serving on " + port)

	server.ListenAndServe()
}

func apiRoute(r chi.Router) {

	r.Use(jsonResponse) // adds json content header

	r.Route("/", PublicRoutes)
	r.Route("/users", UsersRoutes)

	r.Group(func(r chi.Router) {
		r.Use(authentication) // authentication middleware
		r.Route("/scripts", ScriptRoute)
		r.Route("/history", HistoryRoute)
		r.Route("/scripts/{scriptID}/versions", ScriptVersionsRoute)
		r.Route("/scripts/{scriptID}/jobs", JobsRoute)
	})
}

func wsRoute(r chi.Router) {
	r.Use(authentication)
	r.Get("/backtest", backtest)
}

func migrate() error {
	return DB.Gorm().AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.ScriptJob{},
		&db.ScriptTrade{},
		&db.ScriptLog{},
		&db.PendingUser{},
		&db.Credential{},
		&db.User{})
}

func initDB() *db.DB {
	db, err := db.Open(util.PostgresHost(), util.PostgresUser(), util.PostgresDBName(), util.PostgresPassword())
	if err != nil {
		// logger.Fatal(errors.Wrap(err, "failed initializing db"))
	}
	return db
}

func initRedis() *redis.Handle {
	return &redis.Handle{
		Host:        util.RedisHost(),
		MaxActive:   util.RedisMaxActive(),
		MaxIdle:     util.RedisMaxIdle(),
		IdleTimeout: time.Minute * 5,
	}
}
