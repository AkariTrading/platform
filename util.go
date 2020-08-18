package main

import (
	"encoding/json"
	"net/http"
)

func writeJson(w http.ResponseWriter, val interface{}) {

	js, err := json.Marshal(val)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(js)
}
