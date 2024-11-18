package auth

import (
	"context"
	"net/http"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type AuthController struct {
	authService models.AuthServiceInterface
}

func NewAuthController(authService models.AuthServiceInterface) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Auth handles the nickname-hashed-passphrase common dual input and tries to authenticate the user.
//
// @Summary 		Auth an user
// @Description		auth an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param    	 	request body auth.AuthUser true "user struct to auth"
// @Success		200	{object}	common.APIResponse{data=auth.Auth.responseData}
// @Failure		400	{object}	common.APIResponse{data=auth.Auth.responseData}
// @Failure		404	{object}	common.APIResponse{data=auth.Auth.responseData}
// @Failure		500	{object}	common.APIResponse{data=auth.Auth.responseData}
// @Router		/auth [post]
func (c *AuthController) Auth(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "authController")

	// Response body structure.
	type responseData struct {
		AuthGranted bool         `json:"auth_granted"`
		User        *models.User `json:"user"`
	}

	// Prepare the response payload.
	pl := &responseData{
		AuthGranted: false,
		User:        nil,
	}

	var user AuthUser

	// Decode the request body.
	if err := common.UnmarshalRequestData(r, &user); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(pl).Write(w)
		return
	}

	// Try to authenticate given user.
	grantedUser, tokens, err := c.authService.Auth(r.Context(), &user)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(pl).Write(w)
		return
	}

	pl.AuthGranted = true

	// Compose the access HTTP cookie and set it.
	http.SetCookie(w, &http.Cookie{
		Name:     ACCESS_TOKEN,
		Value:    tokens[0],
		Expires:  time.Now().Add(time.Minute * 15),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	// Compose the refresh HTTP cookie and set it.
	http.SetCookie(w, &http.Cookie{
		Name:     REFRESH_TOKEN,
		Value:    tokens[1],
		Expires:  time.Now().Add(common.TOKEN_TTL),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteDefaultMode,
	})

	// Flush the sensitive data of such user in the response.
	patchedUser := (*common.FlushUserData(&map[string]models.User{grantedUser.Nickname: *grantedUser}, grantedUser.Nickname))[grantedUser.Nickname]
	pl.User = &patchedUser
	pl.AuthGranted = true

	l.Msg("ok, auth granted, sending cookies").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// Logout send a client invalidated cookies to cease the session created before.
//
// @Summary 		Log-out an user
// @Description		log-out an user
// @Tags		auth
// @Accept		json
// @Produce		json
// @Success		200	{object}	common.APIResponse{data=auth.Logout.responseData}
// @Router		/auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "authController")

	// Response body structure.
	type responseData struct {
		AuthGranted bool
	}

	// Prepare the response payload.
	pl := &responseData{
		AuthGranted: false,
	}

	cookie, err := r.Cookie(REFRESH_TOKEN)
	if err == nil {
		// Update context with necessary data for the auth service.
		ctx := context.WithValue(r.Context(), "refreshCookie", cookie)

		// Call the auth service to delete the main session (refresh) token.
		if err := c.authService.Logout(ctx); err != nil {
			l.Msg("could not delete associated token").Error(err).Status(http.StatusInternalServerError).Log()
		}
	}

	// Invalidate the access HTTP cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     ACCESS_TOKEN,
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	// Invalidate the refresh HTTP cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     REFRESH_TOKEN,
		Value:    "",
		Expires:  time.Now().Add(time.Second * -300),
		MaxAge:   0,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
	})

	l.Msg("session terminated, void cookies sent (logout)").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
