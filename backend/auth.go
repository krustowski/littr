package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/golang-jwt/jwt"
)

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "auth")
	resp.AuthGranted = false

	var user models.User

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err = json.Unmarshal(data, &user); err != nil {
		resp.Message = "backend error: cannot unmarshall request data"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	l.CallerID = user.Nickname

	// try to authenticate given user
	u, ok := authUser(user)
	if !ok {
		resp.Message = "user not found, or wrong passphrase entered"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// ok, auth granted so far
	// let us generate a JWT
	// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8
	userClaims := UserClaims{
		Nickname:  u.Nickname,
		AppBgMode: u.AppBgMode,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	signedAccessToken, err := NewAccessToken(userClaims)
	if err != nil {
		resp.Message = "error when generating the access token occured"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	signedRefreshToken, err := NewRefreshToken(refreshClaims)
	if err != nil {
		resp.Message = "error when generating the refresh token occured"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// compose cookies
	accessCookie := &http.Cookie{
		Name:    "access-token",
		Value:   signedAccessToken,
		Expires: time.Now().Add(time.Minute * 15),
		Path:    "/",
	}
	refreshCookie := &http.Cookie{
		Name:    "refresh-token",
		Value:   signedRefreshToken,
		Expires: time.Now().Add(time.Hour * 168 * 4),
		Path:    "/",
	}

	// save tokens as HTTP-only cookie
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	resp.Users = make(map[string]models.User)
	resp.Users[u.Nickname] = *u
	resp.AuthGranted = ok

	resp.AccessToken = signedAccessToken
	resp.RefreshToken = signedRefreshToken

	resp.Message = "auth granted"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
	return
}

func handleTokenRefresh(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "auth")
	resp.AuthGranted = false

	req := struct{
		AccessToken string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}{}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err = json.Unmarshal(data, &req); err != nil {
		resp.Message = "backend error: cannot unmarshall request data"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message+err.Error(), resp.Code)
		resp.Write(w)
		return
	}

	userClaims := ParseAccessToken(req.AccessToken)
	refreshClaims := ParseRefreshToken(req.RefreshToken)

	// refresh token is expired => user should relogin
	if refreshClaims.Valid() != nil {
		req.RefreshToken, err = NewRefreshToken(*refreshClaims)
		if err != nil {
			log.Fatal("error creating refresh token")
		}
	}

	// access token is expired
	if userClaims.StandardClaims.Valid() != nil && refreshClaims.Valid() == nil {
		req.AccessToken, err = NewAccessToken(*userClaims)
		if err != nil {
			log.Fatal("error creating access token")
		}
	}

	// compose cookies
	accessCookie := &http.Cookie{
		Name:    "access-token",
		Value:   req.AccessToken,
		Expires: time.Now().Add(time.Minute * 15),
		Path:    "/",
		HttpOnly: true,
	}
	refreshCookie := &http.Cookie{
		Name:    "refresh-token",
		Value:   req.RefreshToken,
		Expires: time.Now().Add(time.Hour * 168 * 4),
		Path:    "/",
		HttpOnly: true,
	}

	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	resp.Message = "ok, tokens refreshed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
	return
}

func authUser(aUser models.User) (*models.User, bool) {
	// fetch one user from cache according to the login credential
	user, ok := getOne(UserCache, aUser.Nickname, models.User{})
	if !ok {
		// not found
		return nil, false
	}

	// check the passhash
	if user.Passphrase == aUser.Passphrase {
		// auth granted
		return &user, true
	}

	// auth failed
	return nil, false
}
