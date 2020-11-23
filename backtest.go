package main

import (
	"encoding/json"
	"net/http"

	"github.com/akaritrading/backtest/pkg/backtestclient"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:    1024,
// 	WriteBufferSize:   1024,
// 	EnableCompression: true,
// }

func BacktestRoute(r chi.Router) {
	r.Post("/", backtest)
}

func backtest(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)

	var req backtestclient.BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.ErrorJSON(w, util.ErrorUnkown)
		logger.Error(errors.WithStack(util.ErrorUnkown))
		return
	}

	res, err := backtestClient.NewRequest(r).Backtest(&req)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	util.WriteJSON(w, res)
}

// func backtest(w http.ResponseWriter, r *http.Request) {

// 	logger := middleware.GetLogger(r)

// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		logger.Error(errors.New("failed upgrading to websocket connection"))
// 		return
// 	}
// 	defer conn.Close()

// 	var testrun backtestclient.BacktestRequest
// 	conn.ReadJSON(&testrun)
// 	if err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}

// 	backtest, err := backtestClient.Connect(testrun)
// 	if err != nil {
// 		logger.Error(errors.WithStack(err))
// 		return
// 	}
// 	defer backtest.Close()

// 	for {
// 		t, msg, err := backtest.ReadMessage()
// 		if err != nil {
// 			return
// 		}

// 		err = conn.WriteMessage(t, msg)
// 		if err != nil {
// 			return
// 		}
// 	}
// }
