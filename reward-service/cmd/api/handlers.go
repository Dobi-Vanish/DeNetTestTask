package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"reward-service/data"
	"strconv"
	"time"
)

type UserData struct {
	ID                 int
	RefreshToken       string
	HashedRefreshToken string
	AccessToken        string
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Referrer  string    `json:"referrer,omitempty"`
}

type contextKey string

const userIDKey contextKey = "userID"

// Registrate insert new user to the database
func (app *Config) Registrate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Password  string `json:"password"`
		Active    int    `json:"active"`
		Score     int    `json:"score"`
		Referrer  string `json:"referrer,omitempty"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	user := User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    requestPayload.Active,
		Score:     requestPayload.Score,
		Referrer:  requestPayload.Referrer,
	}
	id, err := app.Repo.Insert(data.User(user))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Succesfully created new user, id: %s", id),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// GetLeaderboard retrieves all users from the database, sort them by points
func (app *Config) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	users, err := app.Repo.GetAll()
	if err != nil {
		app.errorJSON(w, errors.New("couldn't fetch All users"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Fetched all users"),
		Data:    users,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

// Authenticate authenticates user by provided email and password, provides tokens to access
func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Repo.GetByEmail(requestPayload.Email)

	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials 75"), http.StatusBadRequest)
		return
	}

	valid, err := app.Repo.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	secretKey := "some_secret_key"
	userData, err := generateTokens(user.ID, secretKey)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    userData.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	err = validateRefreshToken(userData.HashedRefreshToken, userData.RefreshToken)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// generateToken generates refresh and access tokens for the user
func generateTokens(userID int, secretKey string) (*UserData, error) {
	accessToken, err := generateAccessToken(userID, secretKey)
	if err != nil {
		return nil, err
	}

	refreshToken, hashedRefreshToken, err := generateRefreshToken(secretKey)

	if err != nil {
		return nil, err
	}

	return &UserData{
		ID:                 userID,
		RefreshToken:       refreshToken,
		HashedRefreshToken: hashedRefreshToken,
		AccessToken:        accessToken,
	}, nil
}

// validateRefreshToken validates refresh token for the user
func validateRefreshToken(hashedRefreshToken, refreshToken string) error {
	decodedBytes, err := base64.StdEncoding.DecodeString(hashedRefreshToken)
	err = bcrypt.CompareHashAndPassword(decodedBytes, []byte(refreshToken))
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}
	return nil
}

// generateRefreshToken generates refresh token for the user
func generateRefreshToken(secretKey string) (string, string, error) {
	refreshToken := secretKey

	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return refreshToken, base64.StdEncoding.EncodeToString(hashedRefreshToken), nil
}

// generateAccessToken generates access tokens based on who was authenticated
func generateAccessToken(userID int, secretKey string) (string, error) {
	expirationTime := time.Now().Add(time.Minute * 15)

	claims := &jwt.MapClaims{
		"sub": userID,
		"exp": expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// authTokenMiddleware auths users to get access to some pages only by having access token
func (app *Config) authTokenMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("access_token")
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}
			tokenString := cookie.Value
			claims := &jwt.MapClaims{
				"sub": userIDKey,
			}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				app.errorJSON(w, errors.New("token is not valid"), http.StatusUnauthorized)
				return
			}

			userID, ok := (*claims)["sub"].(float64)
			if !ok {
				app.errorJSON(w, errors.New("invalid token claims"), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, int(userID))
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// addPoint adds some points to some user
func (app *Config) addPoint(point, id int) error {
	err := app.Repo.AddPoints(id, point)
	if err != nil {
		fmt.Println("Error in reward service, couldn't add point")
		return err
	}

	return nil
}

// completeTask completes various task and adding some point to the user
func (app *Config) completeTask(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Points int `json:"points"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("couldnt convert id string to int"), http.StatusBadRequest)
		return
	}
	err = app.addPoint(requestPayload.Points, id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't add points to the user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("complete task worked for user with id %d, added points %d ", id, requestPayload.Points),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// completeTelegramSign completes various task and adding some point to the user
func (app *Config) completeTelegramSign(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't convert id string to int"), http.StatusBadRequest)
		return
	}
	err = app.addPoint(50, id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't add points to the user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("complete telegram sign worked for user with id %d, added points %d ", id, 50),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// completeXSign completes various task and adding some point to the user
func (app *Config) completeXSign(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't convert id string to int"), http.StatusBadRequest)
		return
	}
	err = app.addPoint(75, id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't add points to the user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("complete X sign worked for user with id %d, added points %d ", id, 75),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// retrieveOne retrieves one user from the database by id
func (app *Config) retrieveOne(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("couldnt convert id string to int"), http.StatusBadRequest)
		return
	}
	user, err := app.Repo.GetOne(id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't fetch user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Retrieved one user from the database"),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

// redeemReferrer redeems referrer for the owner of the referrer and for the user, who used it base on id and referrer
func (app *Config) redeemReferrer(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Referrer string `json:"referrer"`
	}
	idStr := chi.URLParam(r, "id")
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("error during converting to string"), http.StatusBadRequest)
		return
	}
	err = app.Repo.RedeemReferrer(id, requestPayload.Referrer)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't redeem referrer"), http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Referrer redeemed"),
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}
