package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

// UsersRoutes -
func UsersRoutes(r chi.Router) {
	r.Post("/", createUser)
	r.Get("/{id}", getUser)
	r.Put("/{id}", updateUser)
	r.Delete("/{id}", deleteUser)
}

func createUser(w http.ResponseWriter, r *http.Request) {

}

func getUser(w http.ResponseWriter, r *http.Request) {

}

func updateUser(w http.ResponseWriter, r *http.Request) {

}

func deleteUser(w http.ResponseWriter, r *http.Request) {

}
