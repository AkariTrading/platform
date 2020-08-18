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

	id, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	script := &db.Script{}

	query := DB.First(script, id)
	if query.Error != nil {
		if query.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, script)
}

func getScripts(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	var scripts []db.Script

	query := DB.Where("user_id = ?", userID).Find(&scripts)
	if query.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, &scripts)
}

func createScript(w http.ResponseWriter, r *http.Request) {

	var script Script
	err := json.NewDecoder(r.Body).Decode(&script)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScript := &db.Script{Title: script.Title, UserID: getUserIDFromContext(r)}

	if err := DB.Create(newScript).Error; err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, newScript)
}

// maybe for updating title?
func updateScript(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	id, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var update db.Script
	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println(userID, id)

	var script db.Script
	query := DB.Where("user_id = ?", userID).First(&script, id)
	if query.Error != nil {
		if query.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	script.Title = update.Title
	if DB.Save(&script).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeJson(w, script)
}

func deleteScript(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)

	id, err := getIDFromURL(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var versions []db.ScriptVersion
	DB.Where("script_id = ?", id).Where("is_running = ?", true).Find(&versions)

	if len(versions) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		writeJson(w, map[string]interface{}{"error": "script_running", "msg": "Script is running. To delete, script must be stopped first"})
		return
	}

	err = DB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Where("user_id = ?", userID).Delete(&db.Script{}, id).Error; err != nil {
			return err
		}

		if err := tx.Where("script_id = ?", id).Delete(&versions).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func getUserIDFromContext(r *http.Request) uint {
	return r.Context().Value(USERID).(uint)
}

func getIDFromURL(r *http.Request) (uint, error) {
	str := chi.URLParam(r, "id")
	num, err := util.StrToUint(str)
	return uint(num), err
}
