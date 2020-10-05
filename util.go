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
