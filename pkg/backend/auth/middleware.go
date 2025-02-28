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
	"go.vxn.dev/littr/pkg/backend/tokens"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/golang-jwt/jwt"
)

// These URL paths are to be skipped by the authentication middleware.
var PathExceptions = []string{
	"/api/v1",
	"/api/v1/auth",
	"/api/v1/auth/logout",
	"/api/v1/dump",
	"/api/v1/live",
	"/api/v1/users/activation",
	"/api/v1/users/passphrase/request",
	"/api/v1/users/passphrase/reset",
}

type responseData struct {
	AuthGranted bool                   `json:"auth_granted"`
	Users       map[string]models.User `json:"users"`
}

var payload *responseData

// The very authentication middleware entrypoint.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ctx := r.Context()

		/*perm, ok := ctx.Value("acl.permission").(YourPermissionType)
		if !ok || !perm.IsAdmin() {
			http.Error(w, http.StatusText(403), 403)
		    	return
		  }*/

		// Skip those HTTP routes.
		if helpers.Contains(PathExceptions, r.URL.Path) ||
			(r.URL.Path == "/api/v1/users" && r.Method == "POST") {
			next.ServeHTTP(w, r)
			return
		}

		// Instantionate the new logger.
		l := common.NewLogger(r, "authMiddleware")

		// Prepare the HTTP response payload.
		payload := &responseData{
			AuthGranted: false,
		}

		// Fetch the server's secret.
		secret := os.Getenv("APP_PEPPER")
		if secret == "" {
			l.Msg("server's secret is not set!").Status(http.StatusInternalServerError).Log().Payload(payload).Write(w)
			return
		}

		var refreshCookie *http.Cookie
		var err error

		// Get the refresh cookie to check its validity.
		if refreshCookie, err = r.Cookie(REFRESH_TOKEN); err != nil {
			l.Msg("client unauthorized (invalid refresh token)").Status(http.StatusUnauthorized).Error(err).Log().Payload(payload).Write(w)
			return
		}

		// Decode the contents of the refresh HTTP cookie, compare the signature with the server's secret.
		refreshClaims := tokens.ParseRefreshToken(refreshCookie.Value, secret)

		// If the refresh token is expired => user should relogin.
		if refreshClaims == nil || refreshClaims.Valid() != nil {
			l.Msg("invalid/expired refresh token").Status(http.StatusUnauthorized).Log().Payload(payload).Write(w)
			return
		}

		var accessCookie *http.Cookie
		var userClaims *tokens.UserClaims

		// Get the access cookie to check its validity.
		if accessCookie, err = r.Cookie(ACCESS_TOKEN); err != nil {
			//l.Msg("invalid/expired access token").Status(http.StatusUnauthorized).Log().Payload(payload).Write(w)
			//return
		} else {
			// Decode the contents of the access HTTP cookie, compare the signature with the server's secret.
			userClaims = tokens.ParseAccessToken(accessCookie.Value, secret)
		}

		// Access cookie is expired (not present), or userClaims can be decoded but the token is invalid (expired).
		// Generate a new access token.
		if err != nil || (userClaims != nil && userClaims.StandardClaims.Valid() != nil) {
			// Fetch the request token's database record
			refToken, err := fetchRefreshTokenRecord(refreshCookie, &w)
			if refToken == nil && err != nil {
				// Token has been invalidated due to its non-existence or invalidity.
				l.Error(err).Status(http.StatusUnauthorized).Log().Payload(payload).Write(w)
				return
			}

			// Invalidate refresh token if the associated user does not exist.
			user, ok := db.GetOne(db.UserCache, refToken.Nickname, models.User{})
			if !ok {
				// Delete such token record form the Token database.
				db.DeleteOne(db.TokenCache, refToken.Hash)
				invalidateRefreshToken(nil, &w)

				l.Msg("referenced user not found in the database").Status(http.StatusUnauthorized).Log().Payload(payload).Write(w)
				return
			}

			// Prepare new access token claims.
			userClaims := tokens.UserClaims{
				Nickname: refToken.Nickname,
				// Set the new access token's validity to 15 minutes only.
				StandardClaims: jwt.StandardClaims{
					IssuedAt:  time.Now().Unix(),
					ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
				},
			}

			// Issue a new access token via the refresh token's validity.
			accessToken, err := tokens.NewAccessToken(userClaims, secret)
			if err != nil {
				l.Msg("access token generation failed").Error(err).Status(http.StatusInternalServerError).Log().Payload(payload).Write(w)
				return
			}

			//
			//  OK, generate and set a new access HTTP cookie.
			//

			// Compose the new access HTTP cookie structure.
			accessCookie := &http.Cookie{
				Name:     ACCESS_TOKEN,
				Value:    accessToken,
				Expires:  time.Now().Add(time.Minute * 15),
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			}

			// Set the access HTTP cookie to response headers.
			http.SetCookie(w, accessCookie)

			// Add the auth-granted user to the response payload.
			payload.Users = make(map[string]models.User)
			payload.Users[refToken.Nickname] = user

			l.Msg("ok, new access token issued").Status(http.StatusOK).Log()

			// Set the HTTP context value for such user's nickname.
			ctx := context.WithValue(r.Context(), "nickname", refToken.Nickname)

			// Register user's activity = refresh the LastActiveTime datetime field.
			noteUsersActivity(refToken.Nickname)
			r = r.WithContext(ctx)

			// Continue with the HTTP request's propragation.
			next.ServeHTTP(w, r)
			return
		}

		//
		//  Check the validity and existence of the refresh token/cookie if the access token is present and valid.
		//

		// Fetch the request token's database record
		refToken, err := fetchRefreshTokenRecord(refreshCookie, &w)
		if refToken == nil && err != nil {
			// Token has been invalidated due to its non-existence or invalidity.
			l.Error(err).Status(http.StatusInternalServerError).Log().Payload(payload).Write(w)
			return
		}

		// Set the HTTP context with the auth-gratend user's nickname.
		ctx := context.WithValue(r.Context(), "nickname", refToken.Nickname)

		// Register the user's activity = refresh the LastTimeActive datetime field.
		noteUsersActivity(refToken.Nickname)
		r = r.WithContext(ctx)

		// Continue with the HTTP request's propragation.
		next.ServeHTTP(w, r)
		return
	})
}

//
//  Helper functions
//

// invalidateRefreshToken is a helper function to invalidate the refresh HTTP cookie (token). The function sends the invalidated HTTP cookie in the response headers.
func invalidateRefreshToken(l common.Logger, w *http.ResponseWriter) bool {
	// Refresh token is invalid = not found in the Token database => user unauthenticated, invalidate the refresh token.
	voidCookie := &http.Cookie{
		Name:     REFRESH_TOKEN,
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	// Set the invalidated refresh HTTP cookie.
	http.SetCookie(*w, voidCookie)

	// Log the message if the logger is present.
	if l != nil {
		l.Msg("the refresh token has been invalidated, please log-in again").Status(http.StatusUnauthorized).Log().Payload(payload).Write(*w)
	}

	return true
}

// fetchRefreshTokenRecord is a helper function to check the refresh token is valid and exists in the database.
func fetchRefreshTokenRecord(refreshCookie *http.Cookie, w *http.ResponseWriter) (*models.Token, error) {
	// Get the refresh token's fingerprint.
	refreshSum := sha256.New()
	refreshSum.Write([]byte(refreshCookie.Value))
	refreshTokenSum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	// Fetch the refresh token details from the Token database.
	token, found := db.GetOne[models.Token](db.TokenCache, refreshTokenSum, models.Token{})
	if !found && w != nil {
		invalidateRefreshToken(nil, w)
		return nil, fmt.Errorf("refresh token's reference has not been found in the database")
	}

	return &token, nil
}
