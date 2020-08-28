package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
)

type Script struct {
	Title string `json:"title"`
}

type ScriptVersion struct {
	Body string `json:"body"`
}

// ScriptRoute -
func ScriptRoute(r chi.Router) {
	r.Get("/", getScripts)
	r.Post("/", createScript)
	r.Get("/{id}", getScript)
	r.Put("/{id}", updateScript)
	r.Delete("/{id}", deleteScript)

}

func getScript(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "id")
	script := &db.Script{}

	query := DB.Where("id = ?", scriptID).First(script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	util.WriteJSON(w, script)
}

func getScripts(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	var scripts []db.Script

	query := DB.Where("user_id = ?", userID).Find(&scripts)
	if query.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, &scripts)
}

func createScript(w http.ResponseWriter, r *http.Request) {

	var script Script
	err := json.NewDecoder(r.Body).Decode(&script)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := db.Script{Title: script.Title, UserID: getUserIDFromContext(r)}

	fmt.Println(newScript)

	if err := DB.Create(&newScript).Error; err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScript)
}

// maybe for updating title?
func updateScript(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "id")

	var update db.Script
	err := json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var script db.Script
	query := DB.Where("user_id = ? AND id  = ?", userID, scriptID).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	script.Title = update.Title
	if DB.Save(&script).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, script)
}

func deleteScript(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "id")

	var script db.Script
	query := DB.Where("id = ? AND user_id = ?", scriptID, userID).First(&script)
	if err := util.QueryError(w, query); err != nil {
		return
	}

	if script.IsRunning {
		util.ErrorJSON(w, http.StatusBadRequest, util.ScriptRunningError)
		return
	}

	err := DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("id = ? AND user_id = ?", scriptID, userID).Delete(&db.Script{}).Error; err != nil {
			return err
		}

		if err := tx.Where("script_id = ?", scriptID).Delete(&db.ScriptVersion{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getUserIDFromContext(r *http.Request) string {
	return r.Context().Value(USERID).(string)
}
