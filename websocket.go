package main

import (
	"log"
	"net/http"

	"github.com/akaritrading/backtest/pkg/backtestclient"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

var backtestHost = "localhost:9090"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsRoute(r chi.Router) {
	r.Use(authentication)
	r.Get("/backtest", backtest)
}

func backtest(w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	var testrun backtestclient.TestRun
	conn.ReadJSON(&testrun)
	if err != nil {
		log.Fatal(err)
		return
	}

	client := backtestclient.BacktestClient{
		Host: backtestHost,
	}

	backtest, err := client.Connect(testrun)
	if err != nil {
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
