package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/redis"
)

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger := middleware.GetLogger(r)

		// DEBUG
		ctx := context.WithValue(r.Context(), middleware.USERID, "d736b408-aa60-43a3-8daa-d6c21a23c417")
		next.ServeHTTP(w, r.WithContext(ctx))
		return
		// DEBUG

		var sessionToken string

		if token := r.Header.Get(sessionTokenHeader); token != "" {
			sessionToken = token
		} else if c, err := r.Cookie("session_token"); err == nil {
			sessionToken = c.Value
		}

		fmt.Println("token ", sessionToken)

		if sessionToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response, err := redis.String(redisHandle.Do(redis.GetKey, sessionToken))

		fmt.Println(response, err)

		if err != nil {
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), middleware.USERID, response)))
	})
}

func jsonResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if (*r).Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
