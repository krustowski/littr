package push

import (
	"net/http"
	"strings"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"

	chi "github.com/go-chi/chi/v5"
)

const (
	LOGGER_WORKER_NAME string = "pushController"
)

type PushController struct {
	subscriptionService models.SubscriptionServiceInterface
}

func NewPushController(
	subscriptionService models.SubscriptionServiceInterface,
) *PushController {

	if subscriptionService == nil {
		return nil
	}

	return &PushController{
		subscriptionService: subscriptionService,
	}
}

// Create is the handler function to ensure that a sent device has been subscribed to notifications.
//
//	@Summary		Create the notifications subscription
//	@Description		This function call takes in a device specification and creates a new user subscription to webpush notifications.
//	@Tags			push
//	@Accept			json
//	@Produce		json
//	@Param			request	body	models.Device	true	"A device to create the notification subscription for."
//	@Success		201		{object}	common.APIResponse{data=models.Stub}	"The subscription has been created successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized.."
//	@Failure		409		{object}	common.APIResponse{data=models.Stub}	"Conflict: a subscription for such device already exists."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal server problem occured."
//	@Router			/push/subscriptions [post]
func (c *PushController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn models.Device

	// Decode the incoming request's body.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if err := c.subscriptionService.Create(r.Context(), &dtoIn); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
	}

	l.Msg("ok, the notifictions subscription has been created for given device").Status(http.StatusCreated).Log().Payload(nil).Write(w)
}

// Update is the handler function used to update an existing subscription.
//
//	@Summary		Update the notification subscription tag
//	@Description		This function call handles a request to change an user's (caller's) notifications subscription for a device specified by UUID param.
//	@Tags			push
//	@Accept			json
//	@Produce		json
//	@Param			uuid	path	string					true	"An UUID of a device to update."
//	@Param			request	body	push.SubscriptionUpdateRequest		true	"The request's body containing fields to modify."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The subscription has been updated successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"The requested device to update not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occured while the update procedure was processing the data."
//	@Router			/push/subscriptions/{uuid} [patch]
func (c *PushController) Update(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn SubscriptionUpdateRequest

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

	var tagName string

	// Decide the tag to be changed.
	if strings.Contains(r.URL.Path, "mention") {
		tagName = "mention"
	} else if strings.Contains(r.URL.Path, "reply") {
		tagName = "reply"
	}

	if tagName == "" {
		l.Msg("tag to update not specified or is invalid").Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if err := c.subscriptionService.Update(r.Context(), uuid, tagName); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// Delete is the handler function to ensure a given subscription deleted from the database.
//
//	@Summary		Delete a subscription
//	@Description		This function call takes an UUID as parameter to fetch and purge a device associated with such ID from the subscribed devices list.
//	@Tags			push
//	@Produce		json
//	@Param			uuid	path		string	true	"An UUID of a device to delete."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"Requested device has been purged from the subscribed devices list."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"The requested device to delete not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occured while processing the delete request."
//	@Router			/push/subscriptions/{uuid} [delete]
func (c *PushController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// Skip blank callerID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Fetch the param from path.
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// Pass the request to the push notifs service.
	if err := c.subscriptionService.Delete(r.Context(), uuid); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription deleted").Status(http.StatusOK).Log().Payload(nil).Write(w)
}

// SendNotification is the handler function for sending new notification(s).
//
//	@Summary		Send a notification
//	@Description		This function call handles the procedure of a new webpush notification creation and firing.
//	@Tags			push
//	@Produce		json
//	@Param			request	body	push.NotificationRequest	true	"An original post which is to fire a notification to its author."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"The notification has been created and sent to the webpush gateway."
//	@Success		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		404		{object}	common.APIResponse{data=models.Stub}	"The requested device to use not found."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"A serious internal problem occured while preparing the notification for firing."
//	@Router			/push [post]
func (c *PushController) SendNotification(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// skip blank callerID
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var dtoIn NotificationRequest

	// Decode the incoming request's body.
	if err := common.UnmarshalRequestData(r, &dtoIn); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if err := c.subscriptionService.SendNotification(r.Context(), dtoIn.PostID); err != nil {
		l.Msg(err.Error()).Status(common.DecideStatusFromError(err)).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, notification(s) are being sent").Status(http.StatusOK).Log().Payload(nil).Write(w)
}
