package main

import (
	"encoding/json"
	"net/http"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersionsHandle)
	r.Post("/", createScriptVersionHandle)
}

func getScriptVersionsHandle(w http.ResponseWriter, r *http.Request) {

	DB := middleware.GetDB(r)
	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)

	scriptID := getFromURL(r, "scriptID")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	versions, _ := DB.GetScriptVersions(scriptID)

	util.WriteJSON(w, versions)
}

func createScriptVersionHandle(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)
	userID := middleware.GetUserID(r)

	scriptID := getFromURL(r, "scriptID")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	var scriptVersion ScriptVersion
	err := json.NewDecoder(r.Body).Decode(&scriptVersion)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScriptVersion := &db.ScriptVersion{ScriptID: scriptID, Body: scriptVersion.Body}
	if err := DB.Gorm().Create(newScriptVersion).Error; err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScriptVersion)
}
