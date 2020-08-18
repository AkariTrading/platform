package main

import (
	"encoding/json"
	"net/http"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
)

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersions)
	r.Post("/", createScriptVersion)

	r.Get("/{versionId}", getScriptVersion)

	r.Post("/{versionId}/run", runScript)
	r.Post("/{versionId}/stop", stopScript)
}

func getScriptVersion(w http.ResponseWriter, r *http.Request) {
}

func getScriptVersions(w http.ResponseWriter, r *http.Request) {
}

func createScriptVersion(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	scriptID, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// versionID, err := getVersionIDFromURL(r)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	query := DB.Where("user_id = ?", userID).First(&db.Script{}, scriptID)
	if query.Error != nil {
		if query.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var scriptVersion ScriptVersion
	err = json.NewDecoder(r.Body).Decode(&scriptVersion)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := &db.ScriptVersion{ScriptID: scriptID, Body: scriptVersion.Body}

	if err := DB.Create(newScript).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, newScript)

}

func runScript(w http.ResponseWriter, r *http.Request) {
}

func stopScript(w http.ResponseWriter, r *http.Request) {
}

func getVersionIDFromURL(r *http.Request) (uint, error) {
	str := chi.URLParam(r, "versionId")
	num, err := util.StrToUint(str)
	return uint(num), err
}
