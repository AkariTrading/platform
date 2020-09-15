package main

import (
	"net/http"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
)

func JobsRoute(r chi.Router) {
	r.Delete("/{jobID}", stopScriptHandle)
}

func stopScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "scriptID")
	jobID := getFromURL(r, "jobID")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	job, query := DB.GetScriptJob(jobID)
	if err := db.QueryError(w, query); err != nil {
		return
	}

	if !job.IsRunning {
		util.ErrorJSON(w, util.ErrorScriptNotRunning)
		return
	}

	err := engineClient.StopScript(job.NodeIP, jobID)
	if err != nil {
		util.ErrorJSON(w, err)
		return
	}
}
