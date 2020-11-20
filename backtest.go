package main

import (
	"net/http"

	"github.com/akaritrading/backtest/pkg/backtestclient"
	"github.com/akaritrading/libs/flag"
	"github.com/akaritrading/libs/middleware"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var backtestClient = backtestclient.BacktestClient{
	Host: flag.BacktestHost(),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func backtest(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(errors.New("failed upgrading to websocket connection"))
		return
	}
	defer conn.Close()

	var testrun backtestclient.Backtest
	conn.ReadJSON(&testrun)
	if err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	backtest, err := backtestClient.Connect(testrun)
	if err != nil {
		logger.Error(errors.WithStack(err))
		return
	}
	defer backtest.Close()

	for {
		t, msg, err := backtest.ReadMessage()
		if err != nil {
			return
		}

		err = conn.WriteMessage(t, msg)
		if err != nil {
			return
		}
	}
}
