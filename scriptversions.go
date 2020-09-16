package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/akaritrading/engine/pkg/engineclient"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// ScriptVersionsRoute -
func ScriptVersionsRoute(r chi.Router) {
	r.Get("/", getScriptVersionsHandle)
	r.Post("/", createScriptVersionHandle)

	r.Post("/{versionId}/run", runScriptHandle)
}

func getScriptVersionsHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "scriptID")
	userID := getUserIDFromContext(r)

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	versions, _ := DB.GetScriptVersions(scriptID)

	util.WriteJSON(w, versions)
}

func createScriptVersionHandle(w http.ResponseWriter, r *http.Request) {

	scriptID := getFromURL(r, "scriptID")
	userID := getUserIDFromContext(r)

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	var scriptVersion ScriptVersion
	err := json.NewDecoder(r.Body).Decode(&scriptVersion)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	newScriptVersion := &db.ScriptVersion{ScriptID: scriptID, Body: scriptVersion.Body}
	if err := DB.Gorm().Create(newScriptVersion).Error; err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, newScriptVersion)
}

func runScriptHandle(w http.ResponseWriter, r *http.Request) {

	userID := getUserIDFromContext(r)
	scriptID := getFromURL(r, "scriptID")
	versionID := getFromURL(r, "versionId")

	_, query := DB.GetScript(userID, scriptID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetScriptVersion(versionID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	_, query = DB.GetRunningScriptJobByVersion(scriptID, versionID)
	if query.Error != gorm.ErrRecordNotFound {
		logger.Error(errors.WithStack(util.ErrorScriptRunning))
		util.ErrorJSON(w, util.ErrorScriptRunning)
		return
	}

	jobrequest, err := jobRequest(r.Body, scriptID, versionID, userID)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = engineClient.StartScript(jobrequest)
	if err != nil {
		logger.Error(errors.WithStack(err))
		util.ErrorJSON(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func jobRequest(r io.Reader, scriptID string, versionID string, userID string) (*engineclient.JobRequest, error) {

	var job engineclient.JobRequest
	json.NewDecoder(r).Decode(&job)

	// exchange, symbolA, symbolB, portfolio, type CANNOT be null
	if job.Exchange == "" || job.SymbolA == "" || job.Portfolio == nil {
		return nil, errors.New("missing fields")
	}

	if _, ok := engineclient.ScriptJobs[job.Type]; !ok {
		return nil, errors.New("missing fields")
	}

	job.ScriptID = scriptID
	job.VersionID = versionID
	job.State = make(map[string]interface{})
	job.ID = db.NewUUID()

	return &job, nil
}
