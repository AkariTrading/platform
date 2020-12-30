package routes

// import (
// 	"encoding/json"
// 	"io"
// 	"net/http"
// 	"time"

// 	"github.com/akaritrading/engine/pkg/engineclient"
// 	"github.com/akaritrading/libs/db"
// 	"github.com/akaritrading/libs/errutil"
// 	"github.com/akaritrading/libs/middleware"
// 	"github.com/akaritrading/libs/util"
// 	"github.com/go-chi/chi"
// 	"github.com/pkg/errors"
// 	"gorm.io/gorm"
// )

// func JobsRoute(r chi.Router) {
// 	r.Get("/{jobID}", getJob)
// 	r.Delete("/{jobID}", stopJob)
// 	r.Post("/", runJob)
// 	r.Get("/{jobID}/logs", logs)
// }

// func stopJob(w http.ResponseWriter, r *http.Request) {

// 	logger := middleware.GetLogger(r)
// 	DB := middleware.GetDB(r)
// 	userID := middleware.GetUserID(r)
// 	engineClient := engineclient.GetClient(logger)

// 	jobID := getFromURL(r, "jobID")

// 	job, query := DB.GetJobByUserID(userID, jobID)
// 	if err := db.QueryError(w, query); err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}

// 	if !job.IsRunning {
// 		errutil.ErrorJSON(w, util.ErrorScriptNotRunning)
// 		logger.Error(errors.WithStack(util.ErrorScriptNotRunning))
// 		return
// 	}

// 	err := engineClient.StopScript(job.NodeIP, jobID)
// 	if err != nil {
// 		logger.Error(errors.WithStack(err))
// 		errutil.ErrorJSON(w, err)
// 		return
// 	}
// }

// func runJob(w http.ResponseWriter, r *http.Request) {

// 	logger := middleware.GetLogger(r)
// 	userID := middleware.GetUserID(r)
// 	DB := middleware.GetDB(r)
// 	engineClient := engineclient.GetClient(logger)

// 	newJob, err := jobRequest(r.Body, userID)
// 	if err != nil {
// 		logger.Error(errors.WithStack(err))
// 		errutil.ErrorJSON(w, err)
// 		return
// 	}

// 	// needs a connected exchange if cycle eg: not dry run
// 	if newJob.Type == "cycle" {
// 		_, query := DB.GetConnectedExchange(userID, newJob.ExchangeID)
// 		if query.Error != gorm.ErrRecordNotFound {
// 			logger.Error(errors.WithStack(query.Error))
// 			errutil.ErrorJSON(w, util.ErrorExchangeUserCredsNotFound)
// 			return
// 		}
// 	}

// 	err = engineClient.StartScript(newJob)
// 	if err != nil {
// 		logger.Error(errors.WithStack(err))
// 		errutil.ErrorJSON(w, err)
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	util.WriteJSON(w, newJob)
// }

// func logs(w http.ResponseWriter, r *http.Request) {

// 	DB := middleware.GetDB(r)
// 	logger := middleware.GetLogger(r)
// 	userID := middleware.GetUserID(r)
// 	jobID := getFromURL(r, "jobID")

// 	createdBeforeMs := URLQueryInt(r, "createdBefore")
// 	var createdBefore time.Time
// 	if createdBeforeMs == 0 {
// 		createdBefore = time.Now()
// 	} else {
// 		createdBefore = time.Unix(createdBeforeMs/1000, 0)
// 	}

// 	// createdAfterMs := URLQueryInt(r, "createdAfter")
// 	// createdAfter := time.Unix(createdAfterMs/1000, 0)

// 	limit := URLQueryInt(r, "limit")
// 	if limit == 0 {
// 		limit = 100
// 	}

// 	_, query := DB.GetJobByUserID(userID, jobID)
// 	if err := db.QueryError(w, query); err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}

// 	logs, query := DB.GetJobLogs(jobID, createdBefore, int(limit))
// 	if err := db.QueryError(w, query); err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}

// 	util.WriteJSON(w, logs)
// }

// func getJob(w http.ResponseWriter, r *http.Request) {

// 	DB := middleware.GetDB(r)
// 	logger := middleware.GetLogger(r)
// 	userID := middleware.GetUserID(r)
// 	jobID := getFromURL(r, "jobID")

// 	job, query := DB.GetJobByUserID(userID, jobID)
// 	if err := db.QueryError(w, query); err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}

// 	util.WriteJSON(w, job)
// }

// func jobRequest(r io.Reader, userID string) (*engineclient.JobRequest, error) {

// 	var job engineclient.JobRequest
// 	json.NewDecoder(r).Decode(&job)

// 	// exchange, symbolA, symbolB, portfolio, type CANNOT be null
// 	if job.Exchange == "" || job.SymbolA == "" || job.SymbolB == "" || job.ExchangeID == "" || job.Body == "" {
// 		return nil, errors.New("missing fields")
// 	}

// 	if job.Balance == nil {
// 		job.Balance = map[string]float64{}
// 	}

// 	job.State = make(map[string]interface{})
// 	job.ID = util.CreateID()
// 	job.UserID = userID

// 	return &job, nil
// }
