package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	uuid "github.com/satori/go.uuid"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	sendGridKey = "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"
)

func getFromURL(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// CreateUUID -
func CreateUUID() string {
	return uuid.NewV4().String()
}

// SendEmail -
func SendEmail(targetEmail string, url string) {
	from := mail.NewEmail("Akari Trading Test", "akari.trading.test@gmail.com")
	subject := "Welcome to Akari Trading - Verify Your Account"
	to := mail.NewEmail("AkariTrading Developer", targetEmail)
	plainTextContent := "Welcome to Akari Trading. :happypepe:"
	htmlContent := fmt.Sprintf("<a href='%v'> Verify your email. </a>", url)
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(sendGridKey)
	response, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
