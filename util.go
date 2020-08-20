package main

import (
	"encoding/json"
	"net/http"

	"github.com/jinzhu/gorm"
)

type ErrorType string

const (
	ScriptRunningError ErrorType = "scriptRunning"
)

var httpErrors = map[ErrorType]string{
	ScriptRunningError: "Action is not possible because the script is runnig.",
}

func ErrorJSON(w http.ResponseWriter, errType ErrorType) {
	writeJSON(w, map[string]interface{}{"error": errType, "message": httpErrors[errType]})
}

func queryError(w http.ResponseWriter, db *gorm.DB) error {
	if db.Error != nil {
		if db.RecordNotFound() {
			w.WriteHeader(http.StatusNotFound)
			return db.Error
		}
		w.WriteHeader(http.StatusInternalServerError)
		return db.Error
	}
	return nil
}

func writeJSON(w http.ResponseWriter, val interface{}) error {

	js, err := json.Marshal(val)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Write(js)
	return nil
}
