package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

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

	scriptID := getFromURL(r, "id")

	var script db.Script
	query := DB.Where("id = ? AND user_id = ?", scriptID, getUserIDFromContext(r)).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	var versions []db.ScriptVersion
	query = DB.Where("script_id = ?", scriptID).Find(&versions)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	util.WriteJSON(w, versions)
}

func createScriptVersion(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "id")

	var script db.Script
	query := DB.Where("user_id = ? AND id  = ?", userID, scriptID).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	if script.IsRunning {
		util.ErrorJSON(w, http.StatusBadRequest, util.ScriptRunningError)
		return
	}

	var scriptVersion ScriptVersion
	err := json.NewDecoder(r.Body).Decode(&scriptVersion)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := &db.ScriptVersion{ScriptID: scriptID, Body: scriptVersion.Body}

	if err := DB.Create(newScript).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScript)
}

func runScript(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	versionID := getFromURL(r, "versionId")
	isTest, _ := strconv.ParseBool(chi.URLParam(r, "isTest"))

	fmt.Println(scriptID, versionID)

	var script db.Script
	query := DB.Where("id = ? AND user_id = ?", scriptID, getUserIDFromContext(r)).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	if script.IsRunning {
		util.ErrorJSON(w, http.StatusBadRequest, util.ScriptRunningError)
		return
	}

	fmt.Println("SENDING TO ENGINE")

	body, err := runAtEngine(versionID, isTest)
	if err != nil {
		if body != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(body)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func stopScript(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")

	var script db.Script
	query := DB.Where("id = ? AND user_id = ?", scriptID, getUserIDFromContext(r)).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	if !script.IsRunning {
		util.ErrorJSON(w, http.StatusBadRequest, util.ScriptNotRunningError)
		return
	}

	fmt.Println("SENDING TO ENGINE")

	body, err := stopAtEngine(*script.NodeIP, *script.ScriptVersionID)
	if err != nil {
		if body != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(body)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
