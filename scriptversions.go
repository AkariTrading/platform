package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
)

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersions)
	r.Post("/", createScriptVersion)

	r.Post("/{versionId}/run", runScript)
	r.Post("/{versionId}/stop", stopScript)
}

func getScriptVersions(w http.ResponseWriter, r *http.Request) {

	scriptID, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var versions []db.ScriptVersion

	query := DB.Where("script_id = ?", scriptID).Find(&versions)
	if err := queryError(w, query); err != nil {
		return
	}

	writeJSON(w, versions)
}

func createScriptVersion(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	scriptID, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	query := DB.Where("user_id = ?", userID).First(&db.Script{}, scriptID)
	if err := queryError(w, query); err != nil {
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

	writeJSON(w, newScript)
}

func runScript(w http.ResponseWriter, r *http.Request) {

	scriptID, err := getFromURL(r, "id")
	fmt.Println(scriptID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	versionID, err := getFromURL(r, "versionId")
	fmt.Println(versionID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var script db.Script
	query := DB.First(&script, scriptID)
	if err := queryError(w, query); err != nil {
		return
	}

	if script.IsRunning {
		w.WriteHeader(http.StatusBadRequest)
		ErrorJSON(w, ScriptRunningError)
	}

	var version db.ScriptVersion
	query = DB.First(&version, versionID)
	if err := queryError(w, query); err != nil {
		return
	}

	if runAtEngine(versionID, false) {
		return
	}

	w.WriteHeader(http.StatusInternalServerError)

}

func stopScript(w http.ResponseWriter, r *http.Request) {
}

func getFromURL(r *http.Request, key string) (uint, error) {
	str := chi.URLParam(r, key)
	num, err := util.StrToUint(str)
	return uint(num), err
}
