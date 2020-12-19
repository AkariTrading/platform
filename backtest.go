package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func BacktestRoute(r chi.Router) {
	r.Post("/", backtest)
	r.Post("/", task)
}

func backtest(w http.ResponseWriter, r *http.Request) {
	backtestClient.NewRequest(w, r).ProxyBacktest()
}

func task(w http.ResponseWriter, r *http.Request) {
	backtestClient.NewRequest(w, r).ProxyTask()
}

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:    1024,
// 	WriteBufferSize:   1024,
// 	EnableCompression: true,
// }

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
