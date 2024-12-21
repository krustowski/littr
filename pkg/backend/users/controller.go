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

const (
	LOGGER_WORKER_NAME string = "userController"
)

// Structure contents definition for the controller.
type UserController struct {
	postService models.PostServiceInterface
	statService models.StatServiceInterface
	userService models.UserServiceInterface
}

// NewUserController return a pointer to the new controller instance, that has to be populated with User and Post services.
func NewUserController(
	postService models.PostServiceInterface,
	statService models.StatServiceInterface,
	userService models.UserServiceInterface,
) *UserController {
	if postService == nil || statService == nil || userService == nil {
		return nil
	}

	return &UserController{
		postService: postService,
		statService: statService,
		userService: userService,
	}
}

// Create is the users handler that processes input and creates a new user.
//
//	@Summary		Add new user
//	@Description		This function call provides a method on how to create a new user in the system.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body	users.UserCreateRequest 	true		     "The request body containing all listed fields for a new user's creation."
//	@Success		201		{object}	common.APIResponse{data=models.Stub} "The request was processed successfully and an user was created."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub} "The request body contains invalid data, or data types."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub} "This response code may occur when the registration is disabled."
//	@Failure		409		{object}	common.APIResponse{data=models.Stub} "The nickname and/or e-mail fields contain data, that had been already used by someone else."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub} "There is a problem processing the request in the internal server logic. This may occur when a new user cannot be saved to the database for example."
//	@Router			/users [post]
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserCreateRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Create the user at the UserService.
	if err := c.userService.Create(r.Context(), &DTOIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("new user created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
}

// Activate is a handler function to complete the user's activation procedure.
//
//	@Summary		Activate an user via an UUID string
//	@Description		This function call provides a method for the new user's activation using a received UUID string.
//	@Tags			users
//	@Produce		json
//	@Param			uuid	path		string	true	"The UUID string from the activation e-mail, that is sent to the new user after a successful registration."
//	@Success		200		{object}	common.APIResponse{data=models.Stub} "The user was activated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub} "The request body contains invalid data, or data types."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub} "The UUID string does not match any user in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub} "There is a problem processing the request (e.g. a problem accessing the database)."
//	@Router			/users/activation/{uuid} [post]
func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

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
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Error(err).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("the user has been activated successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Update is the users handler that allows the user to change their lists/options/passphrase.
//
//	@Summary		Update user's list properties
//	@Description		This function call enables the caller to modify lists saved with other user's data in the database.
//	@Description
//	@Description 		Those lists are KV structures, that hold another user's nickname as key, and a boolean as a value to specify whether such list should apply its logic on such user.
//	@Description		At least one list has to be specified.
//	@Tags			users
//	@Produce		json
//	@Param			request	body		users.UserUpdateListsRequest	true	"Lists object data as a desired state recipe."
//	@Param			userID	path		string					true	"ID of the user to update"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"User's lists have been updated."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"This code can occur when one wants to update another user (this feature to be prepared for a possible admin panel function)."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is a processing problem in the internal logic, or some system's component does not behave (e.g. database is unavailable)."
//	@Router			/users/{userID}/lists [patch]
func (c *UserController) UpdateLists(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	const (
		userIDParam     string = "userID"
		updateTypeParam string = "updateType"
	)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, userIDParam)
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserUpdateListsRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(context.WithValue(context.WithValue(r.Context(), updateTypeParam, "lists"), userIDParam, userID), &DTOIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// Update is the users handler that allows the user to change their lists/options/passphrase.
//
//	@Summary		Update user's option properties
//	@Description		This function call enables the caller to modify some of their properties (options) saved in the database.
//	@Description
//	@Description		Note: the duality in the options' configuration (map vs. separated booleans) reflects the attempt for backward compatibility with older clients (v0.45.18 and older).
//	@Description		The preferred one is the map configuration.
//	@Tags			users
//	@Produce		json
//	@Param			request	body		users.UserUpdateOptionsRequest	true	"A JSON object containing at least one option with a desired value."
//	@Param			userID	path		string					true	"ID of the user to update"
//	@Success		200		{object}	common.APIResponse{data=models.Stub} 	"User's options were updated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"Unauthorized attempt to modify a foreign option set."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem (e.g. data could not be saved in database)."
//	@Router			/users/{userID}/options [patch]
func (c *UserController) UpdateOptions(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	const (
		userIDParam     string = "userID"
		updateTypeParam string = "updateType"
	)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, userIDParam)
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserUpdateOptionsRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(context.WithValue(context.WithValue(r.Context(), updateTypeParam, "options"), userIDParam, userID), &DTOIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// Update is the users handler that allows the user to change their lists/options/passphrase.
//
//	@Summary		Update user's passphrase
//	@Description		This function call enables the caller to modify some of their properties saved in the database.
//	@Tags			users
//	@Produce		json
//	@Param			request	body		users.UserUpdatePassphraseRequest	true	"Hexadecimal representation of the sha512-hashed passphrases."
//	@Param			userID	path		string					true	"ID of the user to update"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}
//	@Failure		409		{object}	common.APIResponse{data=models.Stub}
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/{userID}/passphrase [patch]
func (c *UserController) UpdatePassphrase(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	const (
		userIDParam     string = "userID"
		updateTypeParam string = "updateType"
	)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, userIDParam)
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var DTOIn UserUpdatePassphraseRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(context.WithValue(context.WithValue(r.Context(), updateTypeParam, "passphrase"), userIDParam, userID), &DTOIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// UploadAvatar is a handler function to update user's avatar directly in the app.
//
//	@Summary		Post user's avatar
//	@Description	post user's avatar
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		users.UserUploadAvatarRequest	true	"The data object containing the new avatar's data."
//	@Param			userID	path		string					true	"user's ID for avatar update"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/{userID}/avatar [post]
func (c *UserController) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

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

	var DTOIn UserUploadAvatarRequest

	// Decode the incoming request data.
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Call the userService to upload and update the avatar.
	avatarURL, err := c.userService.UpdateAvatar(r.Context(), &DTOIn)
	if err != nil {
		l.Msg(err.Error()).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{
		Key: *avatarURL,
	}

	l.Msg("user's avatar uploaded and updated").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
}

// PassphraseReset handles a new passphrase regeneration.
//
//	@Summary		Reset the passphrase
//	@Description	reset the passphrase
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		users.UserPassphraseResetRequest	true	"fill the e-mail address, or UUID fields"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/passphrase/reset [post]
//	@Router			/users/passphrase/request [post]
func (c *UserController) PassphraseReset(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

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

	var DTOIn UserPassphraseResetRequest

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &DTOIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	err := c.userService.ProcessPassphraseRequest(context.WithValue(r.Context(), "requestType", requestType), &DTOIn)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("passphrase request processed successfully, check your mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Delete is the users handler that processes and deletes given user (oneself) form the database.
//
//	@Summary		Delete user
//	@Description	delete user
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		string	true	"ID of the user to delete"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/{userID} [delete]
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

	err := c.userService.Delete(context.WithValue(r.Context(), "userID", userID), userID)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	l.Msg("user deleted successfully, associated data are being deleted concurrently").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// GetAll is the users handler that processes and returns existing users list.
//
//	@Summary		Get a list of users
//	@Description	get a list of users
//	@Tags			users
//	@Produce		json
//	@Param			X-Page-No	header		string	true	"page number"
//	@Success		200			{object}	common.APIResponse{data=users.GetAll.responseData}
//	@Failure		400			{object}	common.APIResponse{data=models.Stub}
//	@Failure		404			{object}	common.APIResponse{data=models.Stub}
//	@Failure		500			{object}	common.APIResponse{data=models.Stub}
//	@Router			/users [get]
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
		User      models.User                `json:"user"`
		Users     map[string]models.User     `json:"users,omitempty"`
		UserStats map[string]models.UserStat `json:"user_stats,omitempty"`
	}

	// Compose the DTO-out from userService.
	users, err := c.userService.FindAll(context.WithValue(r.Context(), "pageNo", pageNo))
	if err != nil {
		l.Msg("could not fetch all users").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch all users").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	// Omit flowStats and exported users map.
	_, userStats, _, err := c.statService.Calculate(r.Context())
	if err != nil {
		l.Msg("could not fetch user stats").Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg("could not fetch user stats").Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	DTOOut := &responseData{
		User:      (*users)[l.CallerID()],
		Users:     *common.FlushUserData(users, l.CallerID()),
		UserStats: *userStats,
	}

	// Log the message and write the HTTP response.
	l.Msg("listing all users and their stats").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
}

// GetByID is the users handler that processes and returns existing user's details according to callerID.
//
//	@Summary		Get the user's details
//	@Description	get the user's details
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	common.APIResponse{data=users.GetByID.responseData}
//	@Failure		400	{object}	common.APIResponse{data=models.Stub}
//	@Failure		404	{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/caller [get]
//	@Router			/users/{userID} [get]
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
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	// Patch the user's data or whatever this does.
	patchedUser := (*common.FlushUserData(
		&map[string]models.User{
			user.Nickname: *user,
		}, l.CallerID()))[user.Nickname]

	// Prepare the response payload.
	DTOOut := &responseData{
		User:      patchedUser,
		Devices:   nil,
		PublicKey: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Msg("returning fetch user's data").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
}

// GetPosts fetches posts only from specified user.
//
//	@Summary		Get user posts
//	@Description	get user posts
//	@Tags			users
//	@Produce		json
//	@Param			X-Hide-Replies	header		string	false	"hide replies"
//	@Param			X-Page-No		header		string	true	"page number"
//	@Param			userID			path		string	true	"user's ID for their posts"
//	@Success		200				{object}	common.APIResponse{data=users.GetPosts.responseData}
//	@Failure		400				{object}	common.APIResponse{data=models.Stub}
//	@Failure		500				{object}	common.APIResponse{data=models.Stub}
//	@Router			/users/{userID}/posts [get]
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
}
