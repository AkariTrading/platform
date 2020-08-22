package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func getFromURL(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}
