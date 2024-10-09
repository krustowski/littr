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

type RefreshToken string

// authHandler
//
// @Summary 		Auth an user
// @Description		auth an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	common.APIResponse
// @Failure		400	{object}	common.APIResponse
// @Failure		404	{object}	common.APIResponse
// @Failure		500	{object}	common.APIResponse
// @Router		/auth/ [post]
// @security            []
func authHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "auth")

	pl := struct {
		AuthGranted bool                   `json:"auth_granted"`
		Users       map[string]models.User `json:"users"`
	}{
		AuthGranted: false,
	}

	var user models.User

	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(&pl).Write(w)
		return
	}

	l.CallerID = user.Nickname

	// try to authenticate given user
	u, ok := authUser(user)
	if !ok {
		l.Msg(common.ERR_AUTH_FAIL).Status(http.StatusBadRequest).Log().Payload(&pl).Write(w)
		return
	}

	// ok, auth granted so far
	// let us generate a JWT
	// https://pascalallen.medium.com/jwt-authentication-with-go-242215a9b4f8
	secret := os.Getenv("APP_PEPPER")

	userClaims := UserClaims{
		Nickname: u.Nickname,
		//User:     *u,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	signedAccessToken, err := NewAccessToken(userClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_ACC_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(&pl).Write(w)
		return
	}

	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 168 * 4).Unix(),
	}

	signedRefreshToken, err := NewRefreshToken(refreshClaims, secret)
	if err != nil {
		l.Msg(common.ERR_AUTH_REF_TOKEN_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(&pl).Write(w)
		return
	}

	refreshSum := sha256.New()
	refreshSum.Write([]byte(signedRefreshToken))
	sum := fmt.Sprintf("%x", refreshSum.Sum(nil))

	if saved := db.TokenCache.Set(sum, u.Nickname); !saved {
		l.Msg(common.ERR_TOKEN_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(&pl).Write(w)
		return
	}

	// compose cookies
	accessCookie := &http.Cookie{
		Name:     "access-token",
		Value:    signedAccessToken,
		Expires:  time.Now().Add(time.Minute * 15),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	refreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    signedRefreshToken,
		Expires:  time.Now().Add(time.Hour * 168 * 4),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	}

	// save tokens as HTTP-only cookie
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)

	pl.Users = make(map[string]models.User)
	pl.Users[u.Nickname] = *u
	pl.Users = *common.FlushUserData(&pl.Users, u.Nickname)
	pl.AuthGranted = true

	l.Msg("auth granted").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// logoutHandler send a client invalidated cookies to cease the session created before.
//
// @Summary 		Log-out an user
// @Description		log-out an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	common.APIResponse
// @Failure		404	{object}	common.APIResponse
// @Failure		500	{object}	common.APIResponse
// @Router		/auth/logout [post]
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "auth")

	pl := struct {
		AuthGranted bool
	}{
		AuthGranted: false,
	}

	voidAccessCookie := &http.Cookie{
		Name:     "access-token",
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	voidRefreshCookie := &http.Cookie{
		Name:     "refresh-token",
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	}

	http.SetCookie(w, voidAccessCookie)
	http.SetCookie(w, voidRefreshCookie)

	l.Msg("session terminated, void cookies sent (logout)").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}
