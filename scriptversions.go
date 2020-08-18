package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersions)

	r.Get("/{id}", getScriptVersion)

	r.Post("/", createScriptVersion)

	r.Delete("/{id}", deleteScriptVersion)
}

func getScriptVersion(w http.ResponseWriter, r *http.Request) {
}

func getScriptVersions(w http.ResponseWriter, r *http.Request) {
}

func createScriptVersion(w http.ResponseWriter, r *http.Request) {
}

func deleteScriptVersion(w http.ResponseWriter, r *http.Request) {
}
