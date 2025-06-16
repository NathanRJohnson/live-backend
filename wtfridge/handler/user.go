package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/NathanRJohnson/live-backend/wtfridge/model"
	"github.com/NathanRJohnson/live-backend/wtfridge/repository/item"
	"github.com/golang-jwt/jwt/v4"
)

type User struct {
	Repo *item.FirebaseRepo
}

var (
	sessionKey []byte
	refreshKey []byte
)

func (u *User) SetKeys(s string, r string) {
	sessionKey = []byte(s)
	refreshKey = []byte(r)
}

func (u *User) Create(w http.ResponseWriter, r *http.Request) {
	log.Println("Create a user")

	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error unmarshalling request:", err)
		return
	}

	if body.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("username cannot be blank.")
		return
	}

	user := model.User{
		Username: body.Username,
	}

	// TODO: Check for collisions

	err := u.Repo.CreateUser(r.Context(), user)
	if err != nil {
		log.Println("failed to create user.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwt, refresh, err := getSignInTokens(user.Username)
	if err != nil {
		log.Println("failed to generate session tokens:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	tokens := map[string]interface{}{
		"session": jwt,
		"refresh": refresh,
	}

	res, err := json.Marshal(tokens)
	if err != nil {
		log.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func (u *User) Read(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error unmarshalling request:", err)
		return
	}

	if body.Username == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("username cannot be blank.")
		return
	}

	user := model.User{
		Username: body.Username,
	}

	userCollection := u.Repo.GetCollectionRef("USER", nil)
	userDoc := u.Repo.GetDocRef(userCollection, body.Username)
	exists, err := u.Repo.DocExists(r.Context(), userDoc)
	if err != nil {
		log.Println("unable to retrive snapshot from document:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !exists {
		log.Println("user not found")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jwt, refresh, err := getSignInTokens(user.Username)
	if err != nil {
		log.Println("failed to generate session tokens:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tokens := map[string]interface{}{
		"session": jwt,
		"refresh": refresh,
	}

	res, err := json.Marshal(tokens)
	if err != nil {
		log.Println("failed to marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

func (u *User) Refresh(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	session_token, err := getTokenFromHeader(authHeader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// check that the session token is actually expired
	_, err = validateSessionToken(session_token)
	if err == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	var body struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("error unmarshalling request:", err)
		return
	}

	newSessionToken, err := validateRefreshToken(body.RefreshToken)
	if err != nil {
		log.Println("failed to issue new session token:", err)
	}

	jsonToken := map[string]interface{}{
		"session_token": newSessionToken,
	}

	res, err := json.Marshal(jsonToken)
	if err != nil {
		log.Println("failed ot marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func getSignInTokens(username string) (string, string, error) {
	sessionToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	})
	sessionTokenString, err := sessionToken.SignedString(sessionKey)
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 6).Unix(),
	})
	refreshTokenString, err := refreshToken.SignedString(refreshKey)
	if err != nil {
		return "", "", err
	}

	return sessionTokenString, refreshTokenString, nil
}

func validateSessionToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return sessionKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}

func validateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return refreshKey, nil
	})

	if err != nil {
		return "", errors.New("invalid or expired refresh token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		username := claims["username"].(string)
		sessionToken, _, err := getSignInTokens(username)
		if err != nil {
			return "", err
		}
		return sessionToken, nil
	}

	return "", errors.New("could not pull value from refresh token")
}

func getUserClaimsFromHeader(header string) (*Claims, error) {
	token, err := getTokenFromHeader(header)
	if err != nil {
		return nil, err
	}

	return validateSessionToken(token)

}

func getTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", errors.New("auth header not found")
	}

	splits := strings.Split(header, " ")
	if len(splits) != 2 || splits[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}
	return splits[1], nil
}
