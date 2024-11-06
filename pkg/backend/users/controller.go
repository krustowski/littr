package users

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
)

type UserController struct {
	userService models.UserServiceInterface
	postService models.PostServiceInterface
}

func NewUserController(
	postService models.PostServiceInterface,
	userService models.UserServiceInterface,
) *UserController {
	if postService == nil || userService == nil {
		return nil
	}

	return &UserController{
		postService: postService,
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
		l.Msg("could not create a new user").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not create a new user").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("new user created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// Activate is a handler function to complete the user's activation procedure.
//
// @Summary      Activate the user via given UUID
// @Description  activate the user via given UUID
// @Tags         users
// @Produce      json
// @Param        uuid path string true "UUID from the activation mail"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/activation/{uuid} [post]
func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the param value from URL's path.
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_REQUEST_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Activate the user at the userService.
	err := c.userService.Activate(context.WithValue(r.Context(), "uuid", uuid), uuid)
	if err != nil {
		l.Msg("could not activate such user").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not activate such user").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("the user has been activated successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
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
// @Router       /users/{userID}/lists [patch]
// @Router       /users/{userID}/options [patch]
// @Router       /users/{userID}/passphrase [patch]
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
	if updateType == "" {
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
	if err := c.userService.Update(context.WithValue(context.WithValue(r.Context(), "updateType", updateType), "userID", userID), &DTOIn); err != nil {
		l.Msg("could not update the passphrase").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not update the passphrase").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("ok, passphrase updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// UploadAvatar is a handler function to update user's avatar directly in the app.
//
// @Summary      Post user's avatar
// @Description  post user's avatar
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body users.UserUpdateRequest true "new avatar data"
// @Param        userID path string true "user's ID for avatar update"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{userID}/avatar [post]
func (c *UserController) UploadAvatar(w http.ResponseWriter, r *http.Request) {
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

	// Declare the HTTP response data contents type.
	type responseData struct {
		Key string `json:"key"`
	}

	var DTOIn UserUpdateRequest

	// Decode the incoming request data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	// Call the userService to upload and update the avatar.
	avatarURL, err := c.userService.UpdateAvatar(r.Context(), &DTOIn)
	if err != nil {
		l.Msg("could not update user's avatar").Status(http.StatusBadRequest).Error(err).Log()
		l.Msg("could not update user's avatar").Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{
		Key: *avatarURL,
	}

	l.Msg("user's avatar uploaded and updated").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
	return
}

// PassphraseReset handles a new passphrase regeneration.
//
// @Summary      Reset the passphrase
// @Description  reset the passphrase
// @Tags         users
// @Accept       json
// @Produce      json
// @Param    	 request body users.UserUpdateRequest true "fill the e-mail address, or UUID fields"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      404  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/passphrase/reset [post]
// @Router       /users/passphrase/request [post]
func (c *UserController) PassphraseReset(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the userID/nickname from the URI.
	requestType := chi.URLParam(r, "requestType")
	if requestType == "" {
		l.Msg(common.ERR_REQUEST_TYPE_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserUpdateRequest

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	err := c.userService.ProcessPassphraseRequest(context.WithValue(r.Context(), "requestType", requestType), &DTOIn)
	if err != nil {
		l.Msg("could not process the passphrase reset request").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not process the passphrase reset request").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("passphrase request processed successfully, check your mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// Delete is the users handler that processes and deletes given user (oneself) form the database.
//
// @Summary      Delete user
// @Description  delete user
// @Tags         users
// @Produce      json
// @Param        userID path string true "ID of the user to delete"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      403  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /users/{userID} [delete]
func (c *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Take the param from path.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Check for possible user's data forgery attempt.
	if userID != l.CallerID() {
		l.Msg(common.ERR_USER_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	err := c.userService.Delete(r.Context(), userID)
	if err != nil {
		l.Msg("could not delete such user").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not delete such user").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("user deleted successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// GetAll is the users handler that processes and returns existing users list.
//
// @Summary      Get a list of users
// @Description  get a list of users
// @Tags         users
// @Produce      json
// @Param    	 X-Page-No header string true "page number"
// @Success      200  {object}   common.APIResponse{data=users.GetAll.responseData}
// @Failure	 400  {object}   common.APIResponse
// @Failure	 404  {object}   common.APIResponse
// @Failure	 500  {object}   common.APIResponse
// @Router       /users [get]
func (c *UserController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the pageNo from headers.
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	type responseData struct {
		User  models.User            `json:"user"`
		Users map[string]models.User `json:"users,omitempty"`
	}

	// Compose the DTO-out from userService.
	users, err := c.userService.FindAll(context.WithValue(r.Context(), "pageNo", pageNo))
	if err != nil {
		l.Msg("could not fetch all users").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch all users").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{
		User:  (*users)[l.CallerID()],
		Users: *common.FlushUserData(users, l.CallerID()),
	}

	// Log the message and write the HTTP response.
	l.Msg("listing all users").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
	return
}

// GetByID is the users handler that processes and returns existing user's details according to callerID.
//
// @Summary      Get the user's details
// @Description  get the user's details
// @Tags         users
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse{data=users.GetByID.responseData}
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Router       /users/{userID} [get]
func (c *UserController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	type responseData struct {
		User      models.User     `json:"user,omitempty"`
		Devices   []models.Device `json:"devices"`
		PublicKey string          `json:"public_key"`
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if userID == "caller" {
		userID = l.CallerID()
	}

	// Fetch the requested user.
	user, err := c.userService.FindByID(context.WithValue(r.Context(), "userID", userID), userID)
	if err != nil {
		l.Msg("could not fetch requested user").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch requested user").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	patchedUser := (*common.FlushUserData(&map[string]models.User{user.Nickname: *user}, l.CallerID()))[user.Nickname]

	pl := &responseData{
		User:      patchedUser,
		Devices:   nil,
		PublicKey: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Msg("returning fetch user's data").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// GetPosts fetches posts only from specified user.
//
// @Summary      Get user posts
// @Description  get user posts
// @Tags         users
// @Produce      json
// @Param    	 X-Hide-Replies header string false "hide replies"
// @Param    	 X-Page-No header string true "page number"
// @Param        userID path string true "user's ID for their posts"
// @Success      200  {object}  common.APIResponse{data=users.GetPosts.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /users/{userID}/posts [get]
func (c *UserController) GetPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "userController")

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the param from URL path.
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the pageNo from headers.
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log()
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Payload(nil).Write(w)
		return
	}

	ctx := context.WithValue(r.Context(), "pageNo", pageNo)

	// Fetch the optional X-Hide-Replies header's value.
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	ctx = context.WithValue(ctx, "hideReplies", hideReplies)

	// Fetch posts and associated users.
	posts, users, err := c.userService.FindPostsByID(ctx, userID)
	if err != nil {
		l.Msg("could not fetch user's posts").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch user's posts").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	type responseData struct {
		Users map[string]models.User `json:"users"`
		Posts map[string]models.Post `json:"posts"`
		Key   string                 `json:"key"`
	}

	// Prepare the payload.
	pl := &responseData{
		Posts: *posts,
		Users: *users,
		Key:   l.CallerID(),
	}

	l.Msg("listing user's posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
