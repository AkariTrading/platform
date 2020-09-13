package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	r.Get("/{id}", getScriptHandle)
	r.Put("/{id}", updateScriptHandle)
	r.Delete("/{id}", deleteScriptHandle)

}

func getScriptHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	userID := getUserIDFromContext(r)

	var script db.Script
	query := DB.Gorm().Preload("ScriptJob").Where("user_id = ? AND id = ?", userID, scriptID).Take(&script)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	util.WriteJSON(w, script)
}

func getScriptsHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	var scripts []db.Script
	query := DB.Gorm().Preload("ScriptJob").Where("user_id = ?", userID).Find(&scripts)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	util.WriteJSON(w, &scripts)
}

func createScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	var script ScriptRequest
	err := json.NewDecoder(r.Body).Decode(&script)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := db.Script{Title: script.Title, UserID: userID}

	fmt.Println(newScript)

	if err := DB.Gorm().Create(&newScript).Error; err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScript)
}

// maybe for updating title?
func updateScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "id")

	var update db.Script
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	script, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
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
	scriptID := getFromURL(r, "id")

	script, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	_, query = DB.GetScriptJob(scriptID, true)
	if !errors.Is(query.Error, gorm.ErrRecordNotFound) {
		util.ErrorJSON(w, util.ErrorScriptRunning)
		return
	}

	err := DB.Gorm().Select("ScriptVersions", "ScriptJob", "Trades", "ScriptLogs").Delete(&script).Error

	// err := DB.Gorm().Transaction(func(tx *gorm.DB) error {

	// 	db.Select("ScriptVersions", "CreditCards").Delete(&script)
	// 	if err := tx.Where("id = ? AND user_id = ?", scriptID, userID).Delete(&db.Script{}).Error; err != nil {
	// 		return err
	// 	}

	// 	if err := tx.Where("script_id = ?", scriptID).Delete(&db.ScriptVersion{}).Error; err != nil {
	// 		return err
	// 	}

	// 	if err := tx.Where("script_id = ?", scriptID).Delete(&db.ScriptJob{}).Error; err != nil {
	// 		return err
	// 	}

	// 	return nil
	// })

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getUserIDFromContext(r *http.Request) string {
	return r.Context().Value(USERID).(string)
}
