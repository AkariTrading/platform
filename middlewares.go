package main

import (
	"context"
	"net/http"
)

type key int

const USERID key = iota

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), USERID, uint(0)) // here 0 should be replaced with the real user id
		// if user cannot be authenticated, request should exit here with correct status code
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func jsonResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
