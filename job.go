package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func JobsRoute(r chi.Router) {
	r.Delete("/{jobID}", stopScriptHandle)
	r.Post("/", runScriptHandle)
	r.Get("/{jobID}/logs", scriptLogs)
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

	// needs a connected exchange if cycle eg: not dry run
	if newJob.Type == "cycle" {
		_, query := DB.GetConnectedExchange(userID, newJob.ExchangeID)
		if query.Error != gorm.ErrRecordNotFound {
			logger.Error(errors.WithStack(query.Error))
			util.ErrorJSON(w, util.ErrorExchangeUserCredsNotFound)
			return
		}
	}

	_, query := DB.GetScript(userID, newJob.ScriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	jobs, query := DB.GetScriptJobByScriptID(newJob.ScriptID, true)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	if len(jobs) > 0 {
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

	util.WriteJSON(w, newJob)
}

func scriptLogs(w http.ResponseWriter, r *http.Request) {

	DB := middleware.GetDB(r)
	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)
	jobID := getFromURL(r, "jobID")

	createdBeforeMs, _ := strconv.ParseInt(r.URL.Query().Get("createdBefore"), 10, 64)
	var createdBefore time.Time
	if createdBeforeMs == 0 {
		createdBefore = time.Now()
	} else {
		createdBefore = time.Unix(createdBeforeMs/1000, 0)
	}

	createdAfterMs, _ := strconv.ParseInt(r.URL.Query().Get("createdAfter"), 10, 64)
	createdAfter := time.Unix(createdAfterMs/1000, 0)

	limit, _ := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if limit == 0 {
		limit = 100
	}

	fmt.Println(createdBefore, createdAfter)

	job, query := DB.GetScriptJob(jobID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetScript(userID, job.ScriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	logs, query := DB.GetScriptJobLogs(jobID, createdBefore, createdAfter, int(limit))
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	util.WriteJSON(w, logs)
}

func jobRequest(r io.Reader, userID string) (*engineclient.JobRequest, error) {

	var job engineclient.JobRequest
	json.NewDecoder(r).Decode(&job)

	// exchange, symbolA, symbolB, portfolio, type CANNOT be null
	if job.Exchange == "" || job.SymbolA == "" || job.Balance == nil || job.ScriptID == "" || job.ExchangeID == "" {
		return nil, errors.New("missing fields")
	}

	job.State = make(map[string]interface{})
	job.ID = util.CreateID()
	job.UserID = userID

	return &job, nil
}
