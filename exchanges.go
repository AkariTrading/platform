package main

import (
	"encoding/json"
	"net/http"

	"github.com/akaritrading/libs/crypto"
	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/errutil"
	"github.com/akaritrading/libs/exchange"
	"github.com/akaritrading/libs/exchange/binance"
	"github.com/akaritrading/libs/flag"
	"github.com/akaritrading/libs/middleware"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
)

type ExchangeRequest struct {
	Exchange  string
	ApiKey    string
	ApiSecret string
}

func ExchangesRoute(r chi.Router) {
	r.Get("/", getExchanges)
	r.Post("/", connectExchange)
	r.Delete("/{exchangeID}", removeExchange)
}

func getExchanges(w http.ResponseWriter, r *http.Request) {
	DB := middleware.GetDB(r)
	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)

	conn, query := DB.GetConnectedExchanges(userID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
	}

	util.WriteJSON(w, conn)
}

func connectExchange(w http.ResponseWriter, r *http.Request) {

	DB := middleware.GetDB(r)
	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)

	var req ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errutil.ErrorJSON(w, util.ErrorUnkown)
		logger.Error(errors.WithStack(util.ErrorUnkown))
		return
	}

	if req.Exchange != "binance" {
		errutil.ErrorJSON(w, util.ErrorExchangeNotFound)
		return
	}

	if err := testExchange(req); err != nil {
		errutil.ErrorJSON(w, errors.New("could not connect to exchange"))
		return
	}

	encAPIKey, err := crypto.EncryptToBase64(flag.ExchangeKey(), req.ApiKey)
	if err != nil {
		errutil.ErrorJSON(w, util.ErrorUnkown)
		return
	}

	encAPISecret, err := crypto.EncryptToBase64(flag.ExchangeKey(), req.ApiSecret)
	if err != nil {
		errutil.ErrorJSON(w, util.ErrorUnkown)
		return
	}

	exc := &db.ExchangeConnection{
		APIKey:    encAPIKey,
		APISecret: encAPISecret,
		Exchange:  req.Exchange,
		UserID:    userID,
	}

	if err := db.QueryError(w, DB.Gorm().Create(exc)); err != nil {
		logger.Error(errors.WithStack(err))
	}

	util.WriteJSON(w, exc)
}

func testExchange(req ExchangeRequest) error {

	client := &binance.BinanceClient{
		UserClient: exchange.UserClient{ApiKey: req.ApiKey, Secret: req.ApiSecret},
	}

	_, err := client.Account()
	return err
}

// TODO: must check that no script is running with this exchange connection
func removeExchange(w http.ResponseWriter, r *http.Request) {

	DB := middleware.GetDB(r)
	logger := middleware.GetLogger(r)
	userID := middleware.GetUserID(r)
	ID := getFromURL(r, "exchangeID")

	_, query := DB.GetConnectedExchange(userID, ID)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	query = DB.Gorm().Delete(&db.ExchangeConnection{ID: ID})
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
	}
}
