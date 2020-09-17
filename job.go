package main

import (
	"net/http"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

func JobsRoute(r chi.Router) {
	r.Delete("/{jobID}", stopScriptHandle)
}

func stopScriptHandle(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)
	engineClient := engineclient.GetClient(logger)

	scriptID := getFromURL(r, "scriptID")
	jobID := getFromURL(r, "jobID")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	job, query := DB.GetScriptJob(jobID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	if !job.IsRunning {
		util.ErrorJSON(w, util.ErrorScriptNotRunning)
		logger.Error(errors.WithStack(util.ErrorScriptNotRunning))
		return
	}

	err := engineClient.StopScript(job.NodeIP, jobID)
	if err != nil {
		logger.Error(errors.WithStack(err))
		util.ErrorJSON(w, err)
		return
	}
}
