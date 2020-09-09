package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// CredentialModel -
type CredentialModel struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// EmailModel -
type EmailModel struct {
	Email string `json:"email"`
}

// CompleteRegistrationModel -
type CompleteRegistrationModel struct {
	Token string `json:"token"`
}

// PublicRoutes -
func PublicRoutes(r chi.Router) {
	r.Post("/login", login)
	r.Post("/logout", logout)
	r.Post("/verifySession", verifySession)

	r.Post("/register", register)
	r.Post("/resendConfirmationEmail", resendRegistrationEmail)
	r.Get("/completeRegistration", completeRegistration)

	r.Post("/resetPasswordRequest", resetPasswordRequest)
	r.Post("/resetPassword", resetPassword)
}

const (
	pendingUserExpiryInDays = 5
	sessionExpiryInSeconds  = 600
	sessionTokenKey         = "session_token"
)

func login(w http.ResponseWriter, r *http.Request) {

	// get credentials
	var input CredentialModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing creds
	existingCred := &db.Credential{}
	query := DB.Where("email = ?", input.Email).Take(existingCred)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("User not found.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// compare inbound and stored passwords
	err = bcrypt.CompareHashAndPassword([]byte(existingCred.Password), []byte(input.Password))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// store session cookie
	sessionToken := CreateUUID()
	_, err = redisHandle.Do("SETEX", sessionToken, fmt.Sprintf("%v", sessionExpiryInSeconds), input.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    sessionTokenKey,
		Value:   sessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}

func logout(w http.ResponseWriter, r *http.Request) {

	// get session from cookie
	c, err := r.Cookie(sessionTokenKey)
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove session from cache
	response, err := redisHandle.Do("DEL", sessionToken)
	if err != nil {
		fmt.Printf("Error: expiring session_token %v.", sessionToken)
		w.WriteHeader(http.StatusInternalServerError)
	} else if response == 0 {
		fmt.Printf("session_token %v already nil.", sessionToken)
		w.WriteHeader(http.StatusOK)
	} else {
		fmt.Printf("Successfully deleted session_token %v.", sessionToken)
		w.WriteHeader(http.StatusOK)
	}

	// remove session from browser
	cookie := http.Cookie{
		Name:    sessionTokenKey,
		Value:   "",
		Expires: time.Time{},
	}
	http.SetCookie(w, &cookie)
}

func verifySession(w http.ResponseWriter, r *http.Request) {

	// get session from cookie
	c, err := r.Cookie(sessionTokenKey)
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	response, err := redisHandle.Do("GET", sessionToken)
	if err != nil {
		// error fetching from cache
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		// not present in cache
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	util.WriteJSON(w, response)
}

func register(w http.ResponseWriter, r *http.Request) {

	// get credentials
	var input CredentialModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check existing user
	existingCred := &db.User{}
	query := DB.Where("email = ?", input.Email).Take(existingCred)
	if !errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Email already exists.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	existingPendingUser := &db.PendingUser{}
	query = DB.Where("email = ?", input.Email).Take(existingPendingUser)
	if !errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Email already exists as a pending user. Please resend confirmation email.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate password requirements TODO

	// salt and has password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create pending user
	token := CreateUUID()
	newPendingUser := db.PendingUser{
		Email:          input.Email,
		Password:       string(hashPassword),
		ExpirationDate: time.Now().AddDate(0, 0, pendingUserExpiryInDays),
		Token:          token,
	}
	if err := DB.Create(&newPendingUser).Error; err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send registration email
	confirmationURL := fmt.Sprintf("http://localhost:6060/api/completeRegistration?token=%v", token)
	SendEmail(input.Email, confirmationURL)

	w.WriteHeader(http.StatusOK)
}

func resendRegistrationEmail(w http.ResponseWriter, r *http.Request) {

	// get email
	var input EmailModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing pendingUser
	pendingUser := &db.PendingUser{}
	query := DB.Where("email = ?", input.Email).Take(pendingUser)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Pending user does not exist. Please re-register.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// replace previous token
	pendingUser.Token = CreateUUID()
	if DB.Save(&pendingUser).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send registration email
	confirmationURL := fmt.Sprintf("http://localhost:6060/completeRegistration?token=%v", pendingUser.Token)
	SendEmail(input.Email, confirmationURL)
}

func completeRegistration(w http.ResponseWriter, r *http.Request) {

	// get token
	keys, ok := r.URL.Query()["token"]
	if !ok || len(keys[0]) < 1 {
		fmt.Println("Url Param 'token' is missing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token := keys[0]

	// get pendingUser
	pendingUser := &db.PendingUser{}
	query := DB.Where("token = ?", token).Take(pendingUser)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Link is invalid.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if time.Now().After(pendingUser.ExpirationDate) {
		fmt.Println("Link is expired. Please resend confirmation email.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create new user
	err := DB.Transaction(func(tx *gorm.DB) error {
		newCred := db.Credential{Email: pendingUser.Email, Password: pendingUser.Password}
		if err := DB.Create(&newCred).Error; err != nil {
			return err
		}

		newUser := db.User{Email: pendingUser.Email}
		if err := DB.Create(&newUser).Error; err != nil {
			return err
		}

		if err := DB.Where("email = ?", pendingUser.Email).Delete(&db.PendingUser{}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:8080/login?registrationComplete=true", http.StatusSeeOther)
}

// https://security.stackexchange.com/questions/86913/should-password-reset-tokens-be-hashed-when-stored-in-a-database
func resetPasswordRequest(w http.ResponseWriter, r *http.Request) {

	// get email
	var input EmailModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing credential
	cred := &db.Credential{}
	query := DB.Where("email = ?", input.Email).Take(cred)
	if errors.Is(query.Error, gorm.ErrRecordNotFound) {
		fmt.Println("Account does not exist.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// replace previous token
	cred.ResetToken = CreateUUID()
	if DB.Save(&cred).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send resetpassword email
	confirmationURL := fmt.Sprintf("http://localhost/resetPassword?token=%v", cred.ResetToken)
	SendEmail(input.Email, confirmationURL)
}

func resetPassword(w http.ResponseWriter, r *http.Request) {

}
