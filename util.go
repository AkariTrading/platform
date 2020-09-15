package main

import (
	"fmt"
	"net/http"

	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var sendGridKey = util.SendGridKey()

func getFromURL(r *http.Request, key string) string {
	return chi.URLParam(r, key)
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
