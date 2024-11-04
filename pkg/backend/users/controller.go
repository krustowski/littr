package users

import (
	"context"
	"net/http"
	//"strconv"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
)

type UserController struct {
	userService models.UserServiceInterface
}

func NewUserController(userService models.UserServiceInterface) *UserController {
	if userService == nil {
		return nil
	}

	return &UserController{
		userService: userService,
	}
}

// Create is the users handler that processes input and creates a new user.
//
// @Summary      Add new user
// @Description  add new user
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body models.User true "new user's request body"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users [post]
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn models.User

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	// Create the user at the UserService.
	if err := c.userService.Create(r.Context(), &DTOIn); err != nil {
		l.Msg("could not create a new user").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not create a new user").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("new user created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// Update is the users handler that allows the user to change their lists/options/passphrase.
//
// @Summary      Update user's data
// @Description  update user's data
// @Tags         users
// @Produce      json
// @Param    	 request body users.UserUpdateRequest true "data to update"
// @Param        userID path string true "ID of the user to update"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID}/{updateType} [patch]
func (c *UserController) Update(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the request type.
	updateType := chi.URLParam(r, "updateType")
	if userID == "" {
		l.Msg(common.ERR_USER_UPDATE_REQ_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserUpdateRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(context.WithValue(r.Context(), "updateType", updateType), &DTOIn); err != nil {
		l.Msg("could not update the passphrase").Status(decideStatusFromError(err)).Error(err).Log()
		l.Msg("could not update the passphrase").Status(decideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("ok, passphrase updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

func (c *UserController) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) PassphraseResetRequest(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) ResetPassphrase(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

func (c *UserController) GetPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}
}

//
//  Helpers
//

var decideStatusFromError = func(err error) int {
	// HTTP 200 condition.
	if err == nil {
		return http.StatusOK
	}

	// HTTP 403 conditions.
	if err.Error() == common.ERR_POLL_SELF_VOTE || err.Error() == common.ERR_POLL_EXISTING_VOTE || err.Error() == common.ERR_POLL_INVALID_VOTE_COUNT {
		return http.StatusForbidden
	}

	// HTTP 404 condition.
	if err.Error() == common.ERR_POLL_NOT_FOUND {
		return http.StatusNotFound
	}

	// HTTP 500 as default.
	return http.StatusInternalServerError
}
