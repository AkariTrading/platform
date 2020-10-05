package main

import (
	"net/http"
	"time"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

func TradesRoute(r chi.Router) {
	r.Get("/", getTrades)
}

func getTrades(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)
	userID := middleware.GetUserID(r)

	var createdBefore time.Time
	createdBeforeMs := URLQueryInt(r, "createdBefore")
	if createdBeforeMs == 0 {
		createdBefore = time.Now()
	} else {
		createdBefore = time.Unix(createdBeforeMs/1000, 0)
	}

	trades, query := DB.GetTrades(userID, createdBefore, 50)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	util.WriteJSON(w, trades)
}
