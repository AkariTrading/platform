package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/akaritrading/libs/flag"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		return

		// logger := middleware.GetLogger(r)

		// DEBUG
		// ctx := context.WithValue(r.Context(), middleware.USERID, "d736b408-aa60-43a3-8daa-d6c21a23c417")
		// next.ServeHTTP(w, r.WithContext(ctx))
		// return
		// DEBUG

		// var sessionToken string

		// if token := r.Header.Get(sessionTokenHeader); token != "" {
		// 	sessionToken = token
		// } else if c, err := r.Cookie("session_token"); err == nil {

		// 	fmt.Println(c)
		// 	sessionToken = c.Value
		// }

		// if sessionToken == "" {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// response, err := redis.String(redisHandle.Do(redis.GetKey, sessionToken))

		// if err != nil {
		// 	logger.Error(err)
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// if response == "" {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	return
		// }

		// next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), middleware.USERID, response)))
	})
}

var sendGridClient = sendgrid.NewSendClient(flag.SendGridKey())

func getFromURL(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func URLQueryInt(r *http.Request, key string) int64 {
	num, _ := strconv.ParseInt(r.URL.Query().Get(key), 10, 64)
	return num
}

// SendEmail -
func SendEmail(targetEmail string, url string) error {

	from := mail.NewEmail("Akari Trading Test", "esadakar@gmail.com")
	subject := "Welcome to Akari Trading - Verify Your Account"
	to := mail.NewEmail("AkariTrading Developer", targetEmail)
	plainTextContent := "Welcome to Akari Trading. :happypepe:"
	htmlContent := fmt.Sprintf("<a href='%v'> Verify your email. </a>", url)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	_, err := sendGridClient.Send(message)
	return err
}
