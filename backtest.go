package main

import (
	"net/http"

	"github.com/akaritrading/backtest/pkg/backtestclient"
	"github.com/akaritrading/libs/util"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var backtestClient = backtestclient.BacktestClient{
	Host: util.BacktestHost(),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
}

func backtest(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error(errors.New("failed upgrading to websocket connection"))
		return
	}
	defer conn.Close()

	var testrun backtestclient.TestRun
	conn.ReadJSON(&testrun)
	if err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	// TODO: error check testrun

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
