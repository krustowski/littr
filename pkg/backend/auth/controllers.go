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

type UserAuth struct {
	User     string `json:"user_name"`
	PassHash string `json:"pass_hash"`
}

// authHandler
//
// @Summary 		Auth an user
// @Description		auth an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param    	 	request body models.User true "user struct to auth"
// @Success		200	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		400	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		404	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Failure		500	{object}	common.APIResponse{data=auth.authHandler.responseData}
// @Router		/auth [post]
func authHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "auth")

	type responseData struct {
		AuthGranted bool                   `json:"auth_granted"`
		Users       map[string]models.User `json:"users"`
	}

	// prepare the payload
	pl := &responseData{
		AuthGranted: false,
	}

	var user models.User

	// decode the request body
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(pl).Write(w)
		return
	}

	l.CallerID = user.Nickname

	// try to authenticate given user
	grantedUser, ok := authUser(&user)
	if !ok {
		l.Msg(common.ERR_AUTH_FAIL).Status(http.StatusBadRequest).Log().Payload(pl).Write(w)
		return
	}

	// ok, auth granted so far
	// let us generate a JWT
	// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8
	secret := os.Getenv("APP_PEPPER")

	userClaims := UserClaims{
		Nickname: grantedUser.Nickname,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	// get new access token
	signedAccessToken, err := NewAccessToken(userClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_ACC_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(pl).Write(w)
		return
	}

	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(common.TOKEN_TTL).Unix(),
	}

	// get new refresh token
	signedRefreshToken, err := NewRefreshToken(refreshClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_REF_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(pl).Write(w)
		return
	}

	// prepare the refresh token's hash
	refreshSum := sha256.New()
	refreshSum.Write([]byte(signedRefreshToken))
	refreshTokenSum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	// prepare the refresh token struct
	token := models.Token{
		Hash:      refreshTokenSum,
		Nickname:  grantedUser.Nickname,
		CreatedAt: time.Now(),
		TTL:       common.TOKEN_TTL,
	}

	// save new refresh token's hash to the database
	if saved := db.TokenCache.Set(refreshTokenSum, token); !saved {
		l.Msg(common.ERR_TOKEN_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(pl).Write(w)
		return
	}

	// compose the access cookie
	accessCookie := &http.Cookie{
		Name:     "access-token",
		Value:    signedAccessToken,
		Expires:  time.Now().Add(time.Minute * 15),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	// compose the refresh cookie
	refreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    signedRefreshToken,
		Expires:  time.Now().Add(common.TOKEN_TTL),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	// set cookies to the response header
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	// update the payload
	users := make(map[string]models.User)
	users[grantedUser.Nickname] = *grantedUser

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

	type responseData struct {
		AuthGranted bool
	}

	// prepare the payload
	pl := &responseData{
		AuthGranted: false,
	}

	// invalidate the access cookie
	voidAccessCookie := &http.Cookie{
		Name:     "access-token",
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	// invalidate the refresh cookie
	voidRefreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	// set void cookies
	http.SetCookie(w, voidAccessCookie)
	http.SetCookie(w, voidRefreshCookie)

	l.Msg("session terminated, void cookies sent (logout)").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
