package users

import (
	"net/http"
	"os"
	"strconv"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/posts"
	"go.vxn.dev/littr/pkg/models"

	"github.com/go-chi/chi/v5"
)

const (
	LOGGER_WORKER_NAME string = "userController"

	userIDParam     string = "userID"
	updateTypeParam string = "updateType"
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

	var dtoIn UserCreateRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Create the user at the UserService.
	if err := c.userService.Create(r.Context(), &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("new user created successfully").Status(http.StatusCreated).Log().Payload(nil).Write(w)
}

// Subscribe is the handler function to ensure that a sent device has been subscribed to notifications.
//
//	@Summary		Create the notifications subscription
//	@Description		This function call takes in a device specification and creates a new user subscription to webpush notifications.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path	string		true					"ID of the user to update"
//	@Param			request	body	models.Device	true					"A device to create the notification subscription for."
//	@Success		201		{object}	common.APIResponse{data=models.Stub}	"The subscription has been created successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized.."
//	@Failure		409		{object}	common.APIResponse{data=models.Stub}	"Conflict: a subscription for such device already exists."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal server problem occurred."
//	@Router			/users/{userID}/subscriptions [post]
func (c *UserController) Subscribe(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
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

	if userID != l.CallerID() {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
	}

	var dtoIn models.Device

	// Decode the incoming request's body.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if err := c.userService.Subscribe(r.Context(), &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, the notifictions subscription has been created for given device").Status(http.StatusCreated).Log().Payload(nil).Write(w)
}

// Activate is a handler function to complete the user's activation procedure.
//
//	@Summary		Activate an user via an UUID string
//	@Description		This function call provides a method for the new user's activation using a received UUID string in payload.
//	@Tags			users
//	@Produce		json
//	@Param			request	body		users.UserActivationRequest	true	"A received UUID string to activate the user after the registration."
//	@Success		200		{object}	common.APIResponse{data=models.Stub} "The user was activated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub} "The request body contains invalid data, or data types."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub} "The UUID string does not match any user in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub} "There is a problem processing the request (e.g. a problem accessing the database)."
//	@Router			/users/activation [post]
func (c *UserController) Activate(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn UserActivationRequest

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Activate the user at the userService.
	err := c.userService.Activate(r.Context(), dtoIn.UUID)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Error(err).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("the user has been activated successfully").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// UpdateLists is the users handler that allows the user to change their lists.
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

	const ()

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

	var dtoIn UserUpdateListsRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(r.Context(), userID, "lists", &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// UpdateOptions is the users handler that allows the user to change their options.
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

	var dtoIn UserUpdateOptionsRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(r.Context(), userID, "options", &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// UpdatePassphrase is the users handler that allows the user to change their passphrase.
//
//	@Summary		Update user's passphrase
//	@Description		This function call enables the caller to modify their current passphrase. The current and a new passphrase are to be sent (hashed using sha512 algorithm).
//	@Description
//	@Description		The problem there is on how to fetch the current passphrase. This can be achieved using a web browser in dev tools (F12), where the hash is to be found on the Network card.
//	@Description		Another problem is that the server uses a secret (pepper), that is appended to a passphrase before loading it into the has algorithm. This secret cannot be fetched via API, as it is a sensitive variable saved as the environmental variable where the server is run.
//	@Tags			users
//	@Produce		json
//	@Param			request	body		users.UserUpdatePassphraseRequest	true	"Hexadecimal representation of the sha512-hashed passphrases."
//	@Param			userID	path		string					true	"ID of the user to update"
//	@Success		200		{object}	common.APIResponse{data=models.Stub} 	"User has been updated."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"Unauthorized attempt to modify a forigner's passphrase."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem present (e.g. data could not be saved to the database)."
//	@Router			/users/{userID}/passphrase [patch]
func (c *UserController) UpdatePassphrase(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

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

	var dtoIn UserUpdatePassphraseRequest

	// Decode the incoming data.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// Update the user's data at the UserService.
	if err := c.userService.Update(r.Context(), userID, "passphrase", &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, user updated").Status(http.StatusOK).Log().Payload(nil).Write(w)

}

// UpdateSubscription is the handler function used to update an existing subscription.
//
//	@Summary		Update the notification subscription tag
//	@Description		This function call handles a request to change an user's (caller's) notifications subscription for a device specified by UUID param.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		string					true	"ID of the user to update"
//	@Param			uuid	path	string					true	"An UUID of a device to update."
//	@Param			request	body	users.UserUpdateSubscriptionRequest		true	"The request's body containing fields to modify."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The subscription has been updated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"The requested device to update not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occurred while the update procedure was processing the data."
//	@Router			/users/{userID}/subscriptions/{uuid} [patch]
func (c *UserController) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
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

	if userID != l.CallerID() {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
	}

	var dtoIn UserUpdateSubscriptionRequest

	// Decode the incoming request's body.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}
	// Fetch the UUID param from path.
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if len(dtoIn) == 0 {
		l.Msg("tag to update not specified or is invalid").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if err := c.userService.UpdateSubscriptionTags(r.Context(), uuid, dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// UploadAvatar is a handler function to update user's avatar directly in the app.
//
//	@Summary		Post user's avatar
//	@Description		This function call presents a method to change one's avatar URL property while also uploading a new picture as a profile photo. Binary data and a figure's extension (JPG, JPEG, PNG) has to be encapsulated into the JSON object as base64 formatted text.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		users.UserUploadAvatarRequest	true	"The data object containing the new avatar's data."
//	@Param			userID	path		string					true	"User's ID for avatar update."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The avatar was uploaded and updated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"Unauthorized attempt to modify a forigner's avatar."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem present (e.g. data could not be saved to the database)."
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

// PassphraseResetRequest handles a new passphrase generation request creation.
//
//	@Summary		Request the passphrase reset
//	@Description		This function call is to be used when an user forgets their passphrase and wants a new one. This very call generates a request for such reset only.
//	@Description
//	@Description		Internally, this is a mailing procedure as two mails has to be delivered and the content used with the API/client to successfully reset the passphrase. This call generates the first mail.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		users.UserPassphraseRequest	true	"User's  e-mail address."
//	@Success		200		{object}	common.APIResponse{data=models.Stub} 	"The passphrase reset request was created successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem present (e.g. data could not be saved to the database)."
//	@Router			/users/passphrase/request [post]
func (c *UserController) PassphraseResetRequest(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn UserPassphraseRequest

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	err := c.userService.ProcessPassphraseRequest(r.Context(), "request", &dtoIn)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("passphrase request processed successfully, check your mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// PassphraseReset handles a new passphrase regeneration.
//
//	@Summary		Reset the passphrase
//	@Description		This function call is to be called when an user forgets their passphrase and wants a new one. This very call does the passphrase regeneration under the hood specifically.
//	@Description
//	@Description		Internally, this is a mailing procedure as two mails has to be delivered and the content used with the API/client to successfully reset the passphrase. This call generates the second mail.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			request	body		users.UserPassphraseReset	true	"Received UUID string to confirm the reset."
//	@Success		200		{object}	common.APIResponse{data=models.Stub} 	"The passphrase was changed successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data received."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem present (e.g. data could not be saved to the database)."
//	@Router			/users/passphrase/reset [post]
func (c *UserController) PassphraseReset(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip the blank caller's ID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn UserPassphraseReset

	// decode the incoming data
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	err := c.userService.ProcessPassphraseRequest(r.Context(), "reset", &dtoIn)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("passphrase request processed successfully, check your mail inbox").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Unsubscribe is the handler function to ensure a given subscription deleted from the database.
//
//	@Summary		Delete a subscription
//	@Description		This function call takes an UUID as parameter to fetch and purge a device associated with such ID from the subscribed devices list.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		string	true	"User's ID to update subscription for."
//	@Param			uuid	path		string	true	"An UUID of a device to delete."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"Requested device has been purged from the subscribed devices list."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"The requested device to delete not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occurred while processing the delete request."
//	@Router			/users/{userID}/subscriptions/{uuid} [delete]
func (c *UserController) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
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

	if userID != l.CallerID() {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
	}

	// Fetch the param from path.
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Pass the request to the push notifs service.
	if err := c.userService.Unsubscribe(r.Context(), uuid); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription deleted").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Delete is the users handler that processes and deletes given user (oneself) form the database.
//
//	@Summary		Delete user
//	@Description		This function call ensures a caller's user account is deleted while all posted items (posts and polls) are purged too. Additionally, all associated requests and tokens are deleted as well.
//	@Tags			users
//	@Produce		json
//	@Param			userID	path		string	true	"ID of the user to be purged"
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The submitted user account has been deleted."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"Blocked attempt to cross-delete other (foreign) user account. The userID param has to be equal to the caller's nickname."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"Such user does not exist in the system."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"There is an internal processing problem present (e.g. data could not be saved to the database)."
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

	err := c.userService.Delete(r.Context(), userID)
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
//	@Description		This function call retrieves a paginated list of user accounts. The page number starts at 0 (and is the default value if not provided in a request).
//	@Tags			users
//	@Produce		json
//	@Param			X-Page-No	header		integer	false	"Page number (default is 0)"
//	@Success		200			{object}	common.APIResponse{data=users.GetAll.responseData} 	"Requested page of user accounts returned."
//	@Failure		400			{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		500			{object}	common.APIResponse{data=models.Stub}			"A generic problem in the internal system's logic. See the `message` KV in JSON to gain more information."
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
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log()
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Payload(nil).Write(w)
		//return
		pageNo = 0
	}

	type responseData struct {
		User      models.User                `json:"user"`
		Users     map[string]models.User     `json:"users,omitempty"`
		UserStats map[string]models.UserStat `json:"user_stats,omitempty"`
	}

	svcPayload := &UserPagingRequest{
		PageNo:     pageNo,
		PagingSize: 25,
	}

	// Compose the DTO-out from userService.
	users, err := c.userService.FindAll(r.Context(), svcPayload)
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
		return
	}

	// Omit flowStats and exported users map.
	_, userStats, _, err := c.statService.Calculate(r.Context())
	if err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Error(err).Log()
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Payload(nil).Write(w)
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
//	@Description		This function call retrieves an user's data that is to be specified in the URI path (as `userID` param below in the request section). A special keyword `caller` can be used to retrieve all reasonable data for the user calling the API. The identity is assured using the refresh token, which is encoded into the refresh HTTP cookie.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID			path		string	true	"User's ID to be shown"
//	@Success		200			{object}	common.APIResponse{data=users.GetByID.responseData}	"Requested user's data (may be limited according to the caller)."
//	@Failure		400			{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		404			{object}	common.APIResponse{data=models.Stub}			"User not found in the database."
//	@Failure		500			{object}	common.APIResponse{data=models.Stub}			"A generic problem in the internal system's logic. See the `message` KV in JSON to gain more information."
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
		User      models.User `json:"user,omitempty"`
		PublicKey string      `json:"public_key"`
	}

	// Fetch the userID/nickname from the URI.
	userID := chi.URLParam(r, userIDParam)
	if userID == "" {
		l.Msg(common.ERR_USERID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if userID == "caller" {
		userID = l.CallerID()
	}

	// Fetch the requested user.
	user, err := c.userService.FindByID(r.Context(), userID)
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
		PublicKey: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Msg("returning fetch user's data").Status(http.StatusOK).Log().Payload(DTOOut).Write(w)
}

// GetPosts fetches posts only from specified user.
//
//	@Summary		Get user posts
//	@Description		This function call is a very specific combination of the users' and posts' services. It retrieves a paginated list of posts made by such user. Special restrictions are applied, such as the privacy (private account, which is not followed by the caller is shown blank). If the list is empty, ensure you are following such user/account.
//	@Tags			users
//	@Produce		json
//	@Param			X-Hide-Replies	header		string	false	"Optional boolean specifying the request of so-called root posts (those not being a reply). Default is false."
//	@Param			X-Page-No	header		string	false	"Page number (default is 0)."
//	@Param			userID		path		string	true	"User's ID (usually the nickname)."
//	@Success		200				{object}	common.APIResponse{data=users.GetPosts.responseData}	"A paginated list of the user's posts (special restriction may apply)."
//	@Failure		400				{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		500				{object}	common.APIResponse{data=models.Stub}			"A very internal service's logic problem. See the `message` field to gain more information."
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
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		pageNo = 0
	}

	// Fetch the optional X-Hide-Replies header's value.
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	opts := &posts.PostPagingRequest{
		HideReplies:  hideReplies,
		PageNo:       pageNo,
		PagingSize:   25,
		SingleUser:   true,
		SinglePostID: userID,
	}

	posts, users, err := c.postService.FindAll(r.Context(), opts)
	if err != nil {
		l.Error(err).Log()
		return
	}

	type responseData struct {
		Users map[string]models.User `json:"users"`
		Posts map[string]models.Post `json:"posts"`
		Key   string                 `json:"key"`
	}

	// prepare the payload
	pl := &responseData{
		Posts: *posts,
		Users: *common.FlushUserData(users, l.CallerID()),
		Key:   l.CallerID(),
	}

	l.Msg("ok, listing user's posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
}
