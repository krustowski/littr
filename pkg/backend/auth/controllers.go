package auth

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.savla.dev/littr/models"

	"github.com/golang-jwt/jwt"
)

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

type RefreshToken string

func authUser(aUser models.User) (*models.User, bool) {
	// fetch one user from cache according to the login credential
	user, ok := getOne(UserCache, aUser.Nickname, models.User{})
	if !ok {
		// not found
		return nil, false
	}

	// check the passhash
	if user.Passphrase == aUser.Passphrase || user.PassphraseHex == aUser.PassphraseHex {
		// update user's hexadecimal passphrase form, as the binary form is broken and cannot be used on BE
		if user.PassphraseHex == "" && aUser.PassphraseHex != "" {
			user.PassphraseHex = aUser.PassphraseHex
			_ = setOne(UserCache, user.Nickname, user)
		}

		// auth granted
		return &user, true
	}

	// auth failed
	return nil, false
}

// authHandler
//
// @Summary 		Auth an user
// @Description		auth an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	Response
// @Failure		400	{object}	Response
// @Failure		404	{object}	Response
// @Failure		500	{object}	Response
// @Router		/auth [post]
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

	if err = json.Unmarshal(reqBody, &user); err != nil {
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
	secret := os.Getenv("APP_PEPPER")

	userClaims := UserClaims{
		Nickname: u.Nickname,
		User:     *u,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	signedAccessToken, err := NewAccessToken(userClaims, secret)
	if err != nil {
		resp.Message = "error when generating the access token occurred"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 168 * 4).Unix(),
	}

	signedRefreshToken, err := NewRefreshToken(refreshClaims, secret)
	if err != nil {
		resp.Message = "error when generating the refresh token occurred"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	refreshSum := sha256.New()
	refreshSum.Write([]byte(signedRefreshToken))
	sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	if saved := TokenCache.Set(sum, u.Nickname); !saved {
		resp.Message = "new refresh token couldn't be saved on backend"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// compose cookies
	accessCookie := &http.Cookie{
		Name:     "access-token",
		Value:    signedAccessToken,
		Expires:  time.Now().Add(time.Minute * 15),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    signedRefreshToken,
		Expires:  time.Now().Add(time.Hour * 168 * 4),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	// save tokens as HTTP-only cookie
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	resp.Users = make(map[string]models.User)
	resp.Users[u.Nickname] = *u
	resp.AuthGranted = ok

	//resp.AccessToken = signedAccessToken
	//resp.RefreshToken = signedRefreshToken

	resp.Message = "auth granted"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
	return
}

