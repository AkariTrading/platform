package main

import (
	"context"
	"net/http"
)

type key int

const USERID key = iota

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		response, err := redisHandle.Do("GET", sessionToken)
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
		ctx := context.WithValue(r.Context(), USERID, response)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func jsonResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
