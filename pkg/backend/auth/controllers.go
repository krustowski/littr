package auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"

	"github.com/golang-jwt/jwt"
)

// authHandler handles the nickname-hashed-passphrase common dual input and tries to authenticate the user.
//
// @Summary 		Auth an user
// @Description		auth an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param    	 	request body auth.AuthUser true "user struct to auth"
// @Success		200	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		400	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		404	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		500	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Router		/auth [post]
func authHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "auth")

	// Response body structure.
	type responseData struct {
		AuthGranted bool                   `json:"auth_granted"`
		Users       map[string]models.User `json:"users"`
	}

	// Prepare the response payload.
	pl := &responseData{
		AuthGranted: false,
	}

	var user AuthUser

	// Decode the request body.
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(pl).Write(w)
		return
	}

	// Set the callerID to such user's nickname henceforth.
	l.CallerID = user.Nickname

	// Try to authenticate given user.
	grantedUser, ok := authUser(&user)
	if !ok {
		l.Msg(common.ERR_AUTH_FAIL).Status(http.StatusBadRequest).Log().Payload(pl).Write(w)
		return
	}

	// Check if the user has been activated yet.
	if !grantedUser.Active || !grantedUser.Options["active"] {
		l.Msg(common.ERR_USER_NOT_ACTIVATED).Status(http.StatusBadRequest).Log().Payload(pl).Write(w)
		return
	}

	//
	// OK, auth granted so far. Let us generate a JWT for such user.
	// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8
	//

	secret := os.Getenv("APP_PEPPER")

	// Compose the user's personal (access) token content.
	userClaims := UserClaims{
		Nickname: grantedUser.Nickname,
		// Access token is restricted to 15 minutes of its validity.
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	// Get new access token = sign the access token with the server's secret.
	signedAccessToken, err := NewAccessToken(userClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_ACC_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(pl).Write(w)
		return
	}

	// Compose the user's personal (refresh) token content. Refresh token is restricted (mainly) to 4 weeks of its validity.
	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(common.TOKEN_TTL).Unix(),
	}

	// Get new refresh token = sign the refresh token with the server's secret.
	signedRefreshToken, err := NewRefreshToken(refreshClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_REF_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(pl).Write(w)
		return
	}

	// Prepare the refresh token's hash for the database payload.
	refreshSum := sha256.New()
	refreshSum.Write([]byte(signedRefreshToken))
	refreshTokenSum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	// Prepare the refresh token struct for the database saving.
	token := models.Token{
		Hash:      refreshTokenSum,
		CreatedAt: time.Now(),
		Nickname:  grantedUser.Nickname,
		TTL:       common.TOKEN_TTL,
	}

	// Save new refresh token's hash to the Token database.
	if saved := db.TokenCache.Set(refreshTokenSum, token); !saved {
		l.Msg(common.ERR_TOKEN_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(pl).Write(w)
		return
	}

	// Compose the access HTTP cookie.
	accessCookie := &http.Cookie{
		Name:     ACCESS_TOKEN,
		Value:    signedAccessToken,
		Expires:  time.Now().Add(time.Minute * 15),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	// Compose the refresh HTTP cookie.
	refreshCookie := &http.Cookie{
		Name:     REFRESH_TOKEN,
		Value:    signedRefreshToken,
		Expires:  time.Now().Add(common.TOKEN_TTL),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	// Set cookies to the HTTP response header.
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	// Update the response payload = add the granted user to the Users field.
	users := make(map[string]models.User)
	users[grantedUser.Nickname] = *grantedUser

	// Flush the sensitive data of such user in the response.
	pl.Users = *common.FlushUserData(&users, grantedUser.Nickname)
	pl.AuthGranted = true

	l.Msg("ok, auth granted, sending cookies").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// logoutHandler send a client invalidated cookies to cease the session created before.
//
// @Summary 		Log-out an user
// @Description		log-out an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	common.APIResponse{data=auth.logoutHandler.responseData}
// @Router		/auth/logout [post]
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "auth")

	// Response body structure.
	type responseData struct {
		AuthGranted bool
	}

	// Prepare the response payload.
	pl := &responseData{
		AuthGranted: false,
	}

	// Invalidate the access HTTP cookie.
	voidAccessCookie := &http.Cookie{
		Name:     ACCESS_TOKEN,
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	// Invalidate the refresh HTTP cookie.
	voidRefreshCookie := &http.Cookie{
		Name:     REFRESH_TOKEN,
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	// Set void HTTP cookies.
	http.SetCookie(w, voidAccessCookie)
	http.SetCookie(w, voidRefreshCookie)

	l.Msg("session terminated, void cookies sent (logout)").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
