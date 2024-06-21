package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/backend/db"
	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	"github.com/golang-jwt/jwt"
)

const (
	ACCESS_TOKEN  = "access-token"
	REFRESH_TOKEN = "refresh-token"
)

var pathExceptions []string = []string{
	"/api/v1",
	"/api/v1/auth",
	"/api/v1/auth/logout",
	"/api/v1/dump/",
	"/api/v1/posts/live",
	"/api/v1/users/passphrase",
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		/*perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		  		if !ok || !perm.IsAdmin() {
		    			http.Error(w, http.StatusText(403), 403)
		    			return
		  		}*/

		// skip those routes
		if helpers.Contains(pathExceptions, r.URL.Path) || (r.URL.Path == "/api/v1/users" && r.Method == "POST") {
			next.ServeHTTP(w, r)
			return
		}

		resp := common.Response{}
		l := common.NewLogger(r, "auth")
		resp.AuthGranted = false

		secret := os.Getenv("APP_PEPPER")

		var accessCookie *http.Cookie
		var refreshCookie *http.Cookie
		var user models.User
		var err error

		if refreshCookie, err = r.Cookie(REFRESH_TOKEN); err != nil {
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
		accessCookie, err = r.Cookie(ACCESS_TOKEN)
		if err == nil {
			userClaims = ParseAccessToken(accessCookie.Value, secret)
		}

		// access cookie not present or access token is expired
		if err != nil || (userClaims != nil && userClaims.StandardClaims.Valid() != nil) {
			refreshSum := sha256.New()
			refreshSum.Write([]byte(refreshCookie.Value))
			sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

			rawNick, found := db.TokenCache.Get(sum)
			if !found {
				voidCookie := &http.Cookie{
					Name:     REFRESH_TOKEN,
					Value:    "",
					Expires:  time.Now().Add(time.Second * -30),
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
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
			user, ok = db.GetOne(db.UserCache, nickname, models.User{})
			if !ok {
				db.DeleteOne(db.TokenCache, sum)

				voidCookie := &http.Cookie{
					Name:     REFRESH_TOKEN,
					Value:    "",
					Expires:  time.Now().Add(time.Second * -30),
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
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
				Name:     ACCESS_TOKEN,
				Value:    accessToken,
				Expires:  time.Now().Add(time.Minute * 15),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}

			http.SetCookie(w, accessCookie)

			resp.Users = make(map[string]models.User)
			resp.Users[nickname] = user

			/*resp.Message = "ok, new access token issued"
			resp.Code = http.StatusOK

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return*/
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

		rawNick, found := db.TokenCache.Get(sum)
		if !found {
			voidCookie := &http.Cookie{
				Name:     REFRESH_TOKEN,
				Value:    "",
				Expires:  time.Now().Add(time.Second * -30),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
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
		noteUsersActivity(nickname)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
