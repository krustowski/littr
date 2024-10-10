package auth

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/golang-jwt/jwt"
)

const (
	ACCESS_TOKEN  = "access-token"
	REFRESH_TOKEN = "refresh-token"
)

var pathExceptions []string = []string{
	"/api/v1",
	"/api/v1/auth/",
	"/api/v1/auth/logout",
	"/api/v1/dump/",
	"/api/v1/posts/live",
	"/api/v1/users/passphrase/request",
	"/api/v1/users/passphrase/reset",
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

		l := common.NewLogger(r, "auth")

		pl := struct {
			AuthGranted bool                   `json:"auth_granted"`
			Users       map[string]models.User `json:"users"`
		}{
			AuthGranted: false,
		}

		secret := os.Getenv("APP_PEPPER")

		var accessCookie *http.Cookie
		var refreshCookie *http.Cookie
		var user models.User
		var err error

		if refreshCookie, err = r.Cookie(REFRESH_TOKEN); err != nil {
			l.Msg("client unauthorized").Status(http.StatusUnauthorized).Error(err).Log().Payload(&pl).Write(w)
			return
		}

		// decode the contents of refreshCookie
		refreshClaims := ParseRefreshToken(refreshCookie.Value, secret)

		// refresh token is expired => user should relogin
		if refreshClaims.Valid() != nil {
			l.Msg("invalid/expired refresh token").Status(http.StatusUnauthorized).Log().Payload(&pl).Write(w)
			return
		}

		var userClaims *UserClaims
		accessCookie, err = r.Cookie(ACCESS_TOKEN)
		if err == nil {
			userClaims = ParseAccessToken(accessCookie.Value, secret)
		}

		// access cookie not present or access token is expired
		if err != nil || (userClaims != nil && userClaims.StandardClaims.Valid() != nil) {
			// get refresh token's fingerprint
			refreshSum := sha256.New()
			refreshSum.Write([]byte(refreshCookie.Value))
			sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

			rawNick, found := db.TokenCache.Get(sum)
			if !found {
				voidCookie := &http.Cookie{
					Name:     REFRESH_TOKEN,
					Value:    "",
					Expires:  time.Now().Add(time.Second * -300),
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
				}

				http.SetCookie(w, voidCookie)

				l.Msg("the refresh token has been invalidated, please log-in again").Status(http.StatusUnauthorized).Log().Payload(&pl).Write(w)
				return
			}

			nickname, ok := rawNick.(string)
			if !ok {
				l.Msg("cannot assert data type for token's nickname field").Status(http.StatusInternalServerError).Log().Payload(&pl).Write(w)
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

				l.Msg("referenced user not found in the database").Status(http.StatusUnauthorized).Log().Payload(&pl).Write(w)
				return
			}

			userClaims := UserClaims{
				Nickname: nickname,
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
				},
			}

			// issue a new access token using refresh token's validity
			accessToken, err := NewAccessToken(userClaims, secret)
			if err != nil {
				l.Msg("access token generation failed").Error(err).Status(http.StatusInternalServerError).Log().Payload(&pl).Write(w)
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

			pl.Users = make(map[string]models.User)
			pl.Users[nickname] = user

			/*resp.Message = "ok, new access token issued"
			resp.Code = http.StatusOK

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return*/

			ctx := context.WithValue(r.Context(), "nickname", nickname)
			noteUsersActivity(nickname)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
			return
		}

		// get the refresh token's fingerprint
		refreshSum := sha256.New()
		refreshSum.Write([]byte(refreshCookie.Value))
		sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

		rawNick, found := db.TokenCache.Get(sum)
		if !found {
			voidCookie := &http.Cookie{
				Name:     REFRESH_TOKEN,
				Value:    "",
				Expires:  time.Now().Add(time.Second * -300),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}

			http.SetCookie(w, voidCookie)

			l.Msg("the refresh token has been invalidated").Status(http.StatusUnauthorized).Log().Payload(&pl).Write(w)
			return
		}

		nickname, ok := rawNick.(string)
		if !ok {
			l.Msg("cannot assert data type for token's nickname").Status(http.StatusInternalServerError).Log().Payload(&pl).Write(w)
			return
		}

		ctx := context.WithValue(r.Context(), "nickname", nickname)
		noteUsersActivity(nickname)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
		return
	})
}
