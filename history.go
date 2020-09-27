package main

import (
	"net/http"
	"strconv"

	"github.com/akaritrading/libs/flag"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/akaritrading/prices/pkg/pricesclient"
	"github.com/go-chi/chi"
	chimiddleware "github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
)

var binance = &pricesclient.Client{
	Host:     flag.PricesHost(),
	Exchange: "binance",
}

func HistoryRoute(r chi.Router) {
	r.Use(chimiddleware.Compress(5))
	r.Get("/{exchange}/{symbol}", getHistoryHandle)
}

func getHistoryHandle(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)

	exchange := chi.URLParam(r, "exchange")
	symbol := chi.URLParam(r, "symbol")

	maxSize, _ := strconv.ParseInt(r.URL.Query().Get("maxSize"), 10, 64)
	if maxSize == 0 {
		maxSize = flag.DefaultHistorySampleSize()
	}

	// start, err := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	// end, err := strconv.ParseInt(r.URL.Query().Get("start"), 10, 64)
	// if err != nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }

	if exchange == "binance" {
		hist, err := binance.GetHistory(symbol, 0, maxSize)
		if err != nil {
			logger.Error(errors.WithStack(err))
			util.ErrorJSON(w, err)
			return
		}
		util.WriteJSON(w, hist)
		return
	}

	logger.Error(errors.WithStack(util.ErrorExchangeNotFound))
	util.ErrorJSON(w, util.ErrorExchangeNotFound)
}
