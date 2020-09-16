package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"gorm.io/gorm"
)

type ScriptRequest struct {
	Title string `json:"title"`
}

type ScriptVersion struct {
	Body string `json:"body"`
}

// ScriptRoute -
func ScriptRoute(r chi.Router) {
	r.Get("/", getScriptsHandle)
	r.Post("/", createScriptHandle)
	r.Get("/{scriptID}", getScriptHandle)
	r.Put("/{scriptID}", updateScriptHandle)
	r.Delete("/{scriptID}", deleteScriptHandle)
}

func getScriptHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "scriptID")
	userID := getUserIDFromContext(r)

	script, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	util.WriteJSON(w, script)
}

func getScriptsHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	scripts, query := DB.GetScripts(userID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	util.WriteJSON(w, &scripts)
}

func createScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	var script ScriptRequest
	err := json.NewDecoder(r.Body).Decode(&script)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := db.Script{Title: script.Title, UserID: userID}

	fmt.Println(newScript)

	if err := DB.Gorm().Create(&newScript).Error; err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScript)
}

// maybe for updating title?
func updateScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "scriptID")

	var update db.Script
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	script, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	script.Title = update.Title
	if DB.Gorm().Save(&script).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, script)
}

func deleteScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "scriptID")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetScriptJob(scriptID)
	if query.Error != gorm.ErrRecordNotFound {
		logger.Error(errors.WithStack(query.Error))
		util.ErrorJSON(w, util.ErrorScriptRunning)
		return
	}

	err := DB.Gorm().Transaction(func(tx *gorm.DB) error {

		err := tx.Where("script_id = ?", scriptID).Delete(&db.ScriptJob{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("script_id = ?", scriptID).Delete(&db.ScriptTrade{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("script_id = ?", scriptID).Delete(&db.ScriptLog{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("id = ?", scriptID).Delete(&db.Script{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("id = ?", scriptID).Delete(&db.Script{}).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}
}

func getUserIDFromContext(r *http.Request) string {
	return r.Context().Value(USERID).(string)
}
