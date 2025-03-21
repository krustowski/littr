package auth

import (
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
//	@Summary		Auth an user
//	@Description		This function call acts as a procedure to authenticate an user using their credentials (nickname and hashed passphrase). On success, the pair of HTTP cookies are sent with the API response (`refresh-token` and `access-token`).
//	@Description
//	@Description		The hashed string is a concatenation of user's passphrase and the server pepper/secret, which is then hashed using the SHA-512 algorithm.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		auth.AuthUser	true	"User's credentials to authenticate."
//	@Success		200		{object}	common.APIResponse{data=auth.Auth.responseData}		"Authentication process successful, HTTP cookies sent in response."
//	@Failure		400		{object}	common.APIResponse{data=auth.Logout.responseData}	"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=auth.Logout.responseData}	"User not authenticated, wrong passphrase used, or such account does not exist at all."
//	@Failure		404		{object}	common.APIResponse{data=auth.Logout.responseData}	"User not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=auth.Logout.responseData}	"Internal server problem while processing the request."
//	@Router			/auth [post]
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
}

// Logout send a client invalidated cookies to cease the session created before.
//
//	@Summary		Log-out an user
//	@Description		This function call's purpose is to sent void HTTP cookies to the caller. If the `refresh-token` sent with the request is valid, it is set to be purged from database and therefore cannot be used anymore.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	common.APIResponse{data=auth.Logout.responseData}	"Void cookies sent in response."
//	@Failure		429	{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Router			/auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "authController")

	// Response body structure.
	type responseData struct {
		AuthGranted bool `json:"auth_granted" example:"false"`
	}

	// Prepare the response payload.
	pl := &responseData{
		AuthGranted: false,
	}

	_, err := r.Cookie(REFRESH_TOKEN)
	if err == nil {
		// Update context with necessary data for the auth service.
		//ctx := context.WithValue(r.Context(), "refreshCookie", cookie)

		// Call the auth service to delete the main session (refresh) token.
		if err := c.authService.Logout(r.Context()); err != nil {
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
}
