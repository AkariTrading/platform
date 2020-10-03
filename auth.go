package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/akaritrading/libs/db"
	"github.com/akaritrading/libs/flag"
	"github.com/akaritrading/libs/middleware"
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
func AuthRoutes(r chi.Router) {

	r.Use(jsonResponse)

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
	sessionTokenHeader      = "X-Session-Token"
)

func login(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)

	var input CredentialModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	existingCred, query := DB.GetCredential(input.Email)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	user, query := DB.GetUser(input.Email)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(existingCred.Password), []byte(input.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		logger.Error(errors.WithStack(err))
		return
	}

	sessionToken := util.CreateID()
	_, err = redisHandle.Do(redis.SetKeyExpire, sessionToken, sessionExpiryInSeconds, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionTokenKey,
		Value:    sessionToken,
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.Header().Set(sessionTokenHeader, sessionToken)

	util.WriteJSON(w, user)
}

func logout(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)

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

	logger := middleware.GetLogger(r)

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

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)

	// get credentials
	var input CredentialModel
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
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
	token := util.CreateID()

	newPendingUser := db.PendingUser{
		Email:          input.Email,
		Password:       string(hashPassword),
		ExpirationDate: time.Now().Add(time.Hour * 24 * pendingUserExpiryInDays),
		Token:          token,
	}
	if err := DB.Gorm().Create(&newPendingUser).Error; err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send registration email

	if err := SendEmail(input.Email, fmt.Sprintf("http://%s/auth/completeRegistration?token=%v", "localhost:6000", token)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func resendRegistrationEmail(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)

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
	pendingUser.Token = util.CreateID()
	if DB.Gorm().Save(&pendingUser).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}

	// send registration email
	if err := SendEmail(input.Email, fmt.Sprintf("http://%s/auth/completeRegistration?token=%v", "localhost:6000", pendingUser.Token)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func completeRegistration(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)

	token := r.URL.Query().Get("token")

	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	fmt.Println(token)

	var users []db.PendingUser

	DB.Gorm().Find(&users)

	fmt.Println(users)

	// get pendingUser
	pendingUser, query := DB.GetPendingUserWithToken(token)
	if err := db.QueryError(w, query); err != nil {
		logger.Error(errors.WithStack(err))
		return
	}

	if time.Now().After(pendingUser.ExpirationDate) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// create new user
	err := DB.Gorm().Transaction(func(tx *gorm.DB) error {

		newCred := db.Credential{Email: pendingUser.Email, Password: pendingUser.Password}
		if err := DB.Gorm().Create(&newCred).Error; err != nil {
			return err
		}

		newUser := db.User{Email: pendingUser.Email}
		if err := DB.Gorm().Create(&newUser).Error; err != nil {
			return err
		}

		if err := DB.Gorm().Where("email = ?", pendingUser.Email).Delete(&db.PendingUser{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logger.Error(errors.WithStack(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("http://%s/login?registrationComplete=true", flag.PlatformHost()), http.StatusSeeOther)
}

// https://security.stackexchange.com/questions/86913/should-password-reset-tokens-be-hashed-when-stored-in-a-database
func resetPasswordRequest(w http.ResponseWriter, r *http.Request) {

	logger := middleware.GetLogger(r)
	DB := middleware.GetDB(r)

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
	cred.ResetToken = util.CreateID()
	if DB.Gorm().Save(&cred).Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Error(errors.WithStack(err))
		return
	}

	// send resetpassword email
	confirmationURL := fmt.Sprintf("http://localhost:6000/resetPassword?token=%v", cred.ResetToken)
	SendEmail(input.Email, confirmationURL)
}

func resetPassword(w http.ResponseWriter, r *http.Request) {

}
