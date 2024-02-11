package backend

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

type RefreshToken string

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
		resp.Message = "error when generating the access token occured"
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
		resp.Message = "error when generating the refresh token occured"
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
	}
	refreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    signedRefreshToken,
		Expires:  time.Now().Add(time.Hour * 168 * 4),
		Path:     "/",
		HttpOnly: true,
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

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		/*perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		  		if !ok || !perm.IsAdmin() {
		    			http.Error(w, http.StatusText(403), 403)
		    			return
		  		}*/

		// skip those routes
		if r.URL.Path == "/api" || r.URL.Path == "/api/auth" || r.URL.Path == "/api/dump" {
			next.ServeHTTP(w, r)
			return
		}

		resp := response{}
		l := NewLogger(r, "auth")
		resp.AuthGranted = false

		secret := os.Getenv("APP_PEPPER")

		var accessCookie *http.Cookie
		var refreshCookie *http.Cookie
		var user models.User
		var err error

		if refreshCookie, err = r.Cookie("refresh-token"); err != nil {
			// logout --- missing refresh token
			resp.Message = "client unauthorized"
			resp.Code = http.StatusUnauthorized

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// decode the contents of refreshCookie
		refreshClaims := ParseRefreshToken(refreshCookie.Value, secret)

		// refresh token is expired => user should relogin
		if refreshClaims.Valid() != nil {
			resp.Message = "refresh token expired"
			resp.Code = http.StatusUnauthorized

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		var userClaims *UserClaims
		accessCookie, err = r.Cookie("access-token")
		if err == nil {
			userClaims = ParseAccessToken(accessCookie.Value, secret)
		}

		// access cookie not present or access token is expired
		if err != nil || (userClaims != nil && userClaims.StandardClaims.Valid() != nil) {
			refreshSum := sha256.New()
			refreshSum.Write([]byte(refreshCookie.Value))
			sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

			rawNick, found := TokenCache.Get(sum)
			if !found {
				voidCookie := &http.Cookie{
					Name:     "refresh-token",
					Value:    "",
					Expires:  time.Now().Add(time.Second * 1),
					Path:     "/",
					HttpOnly: true,
				}

				http.SetCookie(w, voidCookie)

				resp.Message = "the refresh token has been invalidated"
				resp.Code = http.StatusUnauthorized

				l.Println(resp.Message, resp.Code)
				resp.Write(w)
				return
			}

			nickname, ok := rawNick.(string)
			if !ok {
				resp.Message = "cannot assert data type for nickname"
				resp.Code = http.StatusInternalServerError

				l.Println(resp.Message, resp.Code)
				resp.Write(w)
				return
			}

			// invalidate refresh token on non-existing user referenced
			user, ok = getOne(UserCache, nickname, models.User{})
			if !ok {
				deleteOne(TokenCache, sum)

				voidCookie := &http.Cookie{
					Name:     "refresh-token",
					Value:    "",
					Expires:  time.Now().Add(time.Second * 1),
					Path:     "/",
					HttpOnly: true,
				}

				http.SetCookie(w, voidCookie)

				resp.Message = "referenced user not found"
				resp.Code = http.StatusUnauthorized

				l.Println(resp.Message, resp.Code)
				resp.Write(w)
				return
			}

			userClaims := UserClaims{
				Nickname: nickname,
				User:     user,
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
				},
			}

			// issue a new access token using refresh token's validity
			accessToken, err := NewAccessToken(userClaims, secret)
			if err != nil {
				resp.Message = "error creating the access token: " + err.Error()
				resp.Code = http.StatusInternalServerError

				l.Println(resp.Message, resp.Code)
				resp.Write(w)
				return
			}

			accessCookie := &http.Cookie{
				Name:     "access-token",
				Value:    accessToken,
				Expires:  time.Now().Add(time.Minute * 15),
				Path:     "/",
				HttpOnly: true,
			}

			http.SetCookie(w, accessCookie)

			resp.Users = make(map[string]models.User)
			resp.Users[nickname] = user

			resp.Message = "ok, new access token issued"
			resp.Code = http.StatusOK

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		/*resp.Users = make(map[string]models.User)
		resp.Users[user.Nickname] = user

		resp.Message = "auth granted"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return*/

		refreshSum := sha256.New()
		refreshSum.Write([]byte(refreshCookie.Value))
		sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

		rawNick, found := TokenCache.Get(sum)
		if !found {
			voidCookie := &http.Cookie{
				Name:     "refresh-token",
				Value:    "",
				Expires:  time.Now().Add(time.Second * 1),
				Path:     "/",
				HttpOnly: true,
			}

			http.SetCookie(w, voidCookie)

			resp.Message = "the refresh token has been invalidated"
			resp.Code = http.StatusUnauthorized

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		nickname, ok := rawNick.(string)
		if !ok {
			resp.Message = "cannot assert data type for nickname"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		ctx := context.WithValue(r.Context(), "nickname", nickname)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
