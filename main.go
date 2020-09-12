package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/redis"
	"github.com/akaritrading/libs/util"
)

var DB *db.DB
var redisHandle *redis.Handle
var port = ":6060"

func main() {

	DB = initDB()
	migrate()

	redisHandle = initRedis()
	redisHandle.Connect()
	engineClient = engineclient.Client{RedisHandle: redisHandle}
	defer redisHandle.Close()

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)     // replace eventually
	r.Use(middleware.DefaultLogger) // replace eventually
	r.Use(middleware.RequestID)     //

	r.Route("/api", apiRoute)
	r.Route("/ws", wsRoute)

	server := &http.Server{
		Addr:    port,
		Handler: r,
	}

	println("Now serving on " + port)

	log.Fatal(server.ListenAndServe())
}

func apiRoute(r chi.Router) {

	r.Use(jsonResponse) // adds json content header

	r.Route("/", PublicRoutes)
	r.Route("/users", UsersRoutes)

	r.Group(func(r chi.Router) {
		r.Use(authentication) // authentication middleware
		r.Route("/scripts", ScriptRoute)
		r.Route("/scripts/{id}/versions", ScriptVersionsRoute)
	})
}

func migrate() error {
	return DB.Gorm().AutoMigrate(
		&db.Script{},
		&db.ScriptVersion{},
		&db.ScriptJob{},
		&db.PendingUser{},
		&db.Credential{},
		&db.User{})
}

func initDB() *db.DB {
	db, err := db.Open(util.PostgresHost(), util.PostgresUser(), util.PostgresDBName(), util.PostgresPassword())
	if err != nil {
		log.Fatal(err)
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
