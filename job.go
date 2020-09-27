package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func JobsRoute(r chi.Router) {
	r.Post("/", runScriptHandle)
	r.Delete("/{jobID}", stopScriptHandle)
}

func stopScriptHandle(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)
	userID := middleware.GetUserID(r)
	engineClient := engineclient.GetClient(logger)

	jobID := getFromURL(r, "jobID")

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

	_, query = DB.GetScript(userID, job.ScriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	err := engineClient.StopScript(job.NodeIP, jobID)
	if err != nil {
		logger.Error(errors.WithStack(err))
		util.ErrorJSON(w, err)
		return
	}
}

func runScriptHandle(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)
	DB := middleware.GetDB(r)
	engineClient := engineclient.GetClient(logger)

	newJob, err := jobRequest(r.Body, userID)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, query := DB.GetScript(userID, newJob.ScriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetScriptVersion(newJob.VersionID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetRunningScriptJobByVersion(newJob.ScriptID, newJob.VersionID)
	if query.Error != gorm.ErrRecordNotFound {
		logger.Error(errors.WithStack(util.ErrorScriptRunning))
		util.ErrorJSON(w, util.ErrorScriptRunning)
		return
	}

	err = engineClient.StartScript(newJob)
	if err != nil {
		logger.Error(errors.WithStack(err))
		util.ErrorJSON(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func jobRequest(r io.Reader, userID string) (*engineclient.JobRequest, error) {

	var job engineclient.JobRequest
	json.NewDecoder(r).Decode(&job)

	// exchange, symbolA, symbolB, portfolio, type CANNOT be null
	if job.Exchange == "" || job.SymbolA == "" || job.Portfolio == nil || job.ScriptID == "" || job.VersionID == "" {
		return nil, errors.New("missing fields")
	}

	if _, ok := engineclient.ScriptJobs[job.Type]; !ok {
		return nil, errors.New("missing fields")
	}

	job.State = make(map[string]interface{})
	job.ID = util.CreateID()

	return &job, nil
}
