package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

// ScriptRoute -
func ScriptRoute(r chi.Router) {
	r.Get("/", getScripts)

	r.Get("/{id}", getScript)

	r.Post("/", createScript)

	r.Delete("/{id}", deleteScript)
}

func getScript(w http.ResponseWriter, r *http.Request) {
}

func getScripts(w http.ResponseWriter, r *http.Request) {
}

func createScript(w http.ResponseWriter, r *http.Request) {
}

func deleteScript(w http.ResponseWriter, r *http.Request) {
}
