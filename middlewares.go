package main

import (
	"context"
	"net/http"

	"github.com/akaritrading/libs/redis"
)

type key int

const USERID key = iota

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// DEBUG
		ctx := context.WithValue(r.Context(), USERID, "d736b408-aa60-43a3-8daa-d6c21a23c417")
		next.ServeHTTP(w, r.WithContext(ctx))
		return
		// DEBUG

		// We can obtain the session token from the requests cookies, which come with every request
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionToken := c.Value

		// We then get the name of the user from our cache, where we set the session token
		response, err := redisHandle.Do(redis.GetKey, sessionToken)
		if err != nil {
			// error fetching from cache
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response == nil {
			// not present in cache
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), USERID, response)))
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
