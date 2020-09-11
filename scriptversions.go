package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
)

var engine = engineclient.Client{
	RedisHandle: redisHandle,
}

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersionsHandle)
	r.Post("/", createScriptVersionHandle)

	r.Post("/{versionId}/run", runScriptHandle)
	r.Post("/{versionId}/stop", stopScriptHandle)
}

func getScriptVersionsHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	userID := getUserIDFromContext(r)

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	versions, query := DB.GetScriptVersions(scriptID)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	util.WriteJSON(w, versions)
}

func createScriptVersionHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")

	var scriptVersion ScriptVersion
	err := json.NewDecoder(r.Body).Decode(&scriptVersion)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScriptVersion := &db.ScriptVersion{ScriptID: scriptID, Body: scriptVersion.Body}
	if err := DB.Gorm().Create(newScriptVersion).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScriptVersion)
}

func runScriptHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	versionID := getFromURL(r, "versionId")
	isTest, _ := strconv.ParseBool(r.URL.Query().Get("isTest"))

	_, query := DB.GetScriptJob(scriptID)
	if !query.RecordNotFound() {
		util.ErrorJSON(w, engineclient.ErrorScriptRunning.Error())
		return
	}

	err := engine.StartScript(versionID, isTest)
	if err != nil {
		util.ErrorJSON(w, err.Error())
		return
	}
}

func stopScriptHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	versionID := getFromURL(r, "versionId")

	job, query := DB.GetScriptJob(scriptID)
	if query.RecordNotFound() {
		util.ErrorJSON(w, engineclient.ErrorScriptNotRunning.Error())
		return
	}

	err := engine.StopScript(*job.NodeIP, versionID)
	if err != nil {
		util.ErrorJSON(w, err.Error())
		return
	}
}
