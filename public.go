package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/redis"
	"github.com/akaritrading/libs/util"
	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	sessionExpiryInSeconds  = int64(time.Hour / time.Second)
	sessionTokenKey         = "session_token"
)

func login(w http.ResponseWriter, r *http.Request) {

	// get credentials
	var input CredentialModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing creds
	existingCred, query := DB.GetCredential(input.Email)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	// compare inbound and stored passwords
	err = bcrypt.CompareHashAndPassword([]byte(existingCred.Password), []byte(input.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Error(errors.WithStack(err))
		return
	}

	// store session cookie
	sessionToken := db.NewUUID()
	_, err = redisHandle.Do(redis.SetKeyExpire, sessionToken, sessionExpiryInSeconds, input.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
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
			logger.Error(errors.WithStack(err))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	// remove session from cache
	response, err := redisHandle.Do(redis.DeleteKey, sessionToken)
	if err != nil {
		fmt.Printf("Error: expiring session_token %v.", sessionToken)
		logger.Error(errors.WithStack(err))
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
			logger.Error(errors.WithStack(err))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	response, err := redisHandle.Do(redis.GetKey, sessionToken)
	if err != nil {
		// error fetching from cache
		logger.Error(errors.WithStack(err))

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
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check existing user
	_, query := DB.GetUser(input.Email)
	if query.Error != gorm.ErrRecordNotFound {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, query = DB.GetPendingUser(input.Email)
	if query.Error != gorm.ErrRecordNotFound {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate password requirements TODO

	// salt and has password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create pending user
	token := db.NewUUID()
	newPendingUser := db.PendingUser{
		Email:          input.Email,
		Password:       string(hashPassword),
		ExpirationDate: time.Now().AddDate(0, 0, pendingUserExpiryInDays),
		Token:          token,
	}
	if err := DB.Gorm().Create(&newPendingUser).Error; err != nil {
		logger.Error(errors.WithStack(err))
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
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing pendingUser
	pendingUser, query := DB.GetPendingUser(input.Email)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	// replace previous token
	pendingUser.Token = db.NewUUID()
	if DB.Gorm().Save(&pendingUser).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}

	// send registration email
	confirmationURL := fmt.Sprintf("http://%s/completeRegistration?token=%v", util.PlatformHost(), pendingUser.Token)
	SendEmail(input.Email, confirmationURL)
}

func completeRegistration(w http.ResponseWriter, r *http.Request) {

	// get token

	token := r.URL.Query().Get("token")

	if token == "" {
		fmt.Println("Url Param 'token' is missing")
		w.WriteHeader(http.StatusBadRequest)
	}

	// get pendingUser
	pendingUser, query := DB.GetPendingUserWithToken(token)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	if time.Now().After(pendingUser.ExpirationDate) {
		fmt.Println("Link is expired. Please resend confirmation email.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create new user

	err := DB.Gorm().Transaction(func(tx *gorm.DB) error {

		newCred := db.Credential{Email: pendingUser.Email, Password: pendingUser.Password}
		if err := DB.Gorm().Create(&newCred).Error; err != nil {
			logger.Error(errors.WithStack(err))
			return err
		}

		newUser := db.User{Email: pendingUser.Email}
		if err := DB.Gorm().Create(&newUser).Error; err != nil {
			logger.Error(errors.WithStack(err))
			return err
		}

		if err := DB.Gorm().Where("email = ?", pendingUser.Email).Delete(&db.PendingUser{}).Error; err != nil {
			logger.Error(errors.WithStack(err))
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
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get existing credential
	cred, query := DB.GetCredential(input.Email)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	// replace previous token
	cred.ResetToken = db.NewUUID()
	if DB.Gorm().Save(&cred).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}

	// send resetpassword email
	confirmationURL := fmt.Sprintf("http://localhost/resetPassword?token=%v", cred.ResetToken)
	SendEmail(input.Email, confirmationURL)
}

func resetPassword(w http.ResponseWriter, r *http.Request) {

}
