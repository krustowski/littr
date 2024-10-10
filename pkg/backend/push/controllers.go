package push

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// fetchVAPIDKey is a helper function to fetch new server VAPID key pair.
//
// @Summary      Get a VAPID key pair
// @Description  get a VAPID key pair
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure 	 401  {object}   common.APIResponse
// @Router       /push/vapid [get]
func fetchVAPIDKey(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if callerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pl := struct {
		Key string `json:"key"`
	}{
		Key: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Msg("ok, sending public VAPID key").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// updateSubscription is the push pkg's handler function used to update an existing subscription.
//
// @Summary      Update the notification subscription tag
// @Description  Update the notification subscription tag
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription/{uuid}/mention [put]
// @Router       /push/subscription/{uuid}/reply [put]
func updateSubscription(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	path := r.URL.Path
	payload := models.Device{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache later on
	devs, ok := db.GetOne(db.SubscriptionCache, callerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_DEVICE_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	tag := ""
	if strings.Contains(path, "mention") {
		tag = "mention"
	} else if strings.Contains(path, "reply") {
		tag = "reply"
	}

	if !helpers.Contains(payload.Tags, tag) {
		payload.Tags = append(payload.Tags, tag)
	} else {
		tags := payload.Tags
		newTags := []string{}
		for _, t := range tags {
			if t == tag {
				continue
			}
			newTags = append(newTags, t)
		}
		payload.Tags = newTags
	}

	// if found, update the existing device
	found := false
	for idx, dev := range devs {
		if dev.UUID == payload.UUID {
			devs[idx] = payload
			found = true
		}
	}

	// append new device into the devices array for such user
	if !found {
		devs = append(devs, payload)
	}

	if saved := db.SetOne(db.SubscriptionCache, callerID, devs); !saved {
		l.Msg(common.ERR_SUBSCRIPTION_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// subscribeToNotifications is the push pkg's handler function to ensure sent device has subscribed to notifications.
//
// @Summary      Add the notification subscription
// @Description  add the notification subscription
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription [post]
func subscribeToNotifications(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	payload := models.Device{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache
	devs, ok := db.GetOne(db.SubscriptionCache, callerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_DEVICE_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	for _, dev := range devs {
		if dev.UUID == payload.UUID {
			// we have just found a match, thus request was fired twice, or someone tickles the API
			l.Msg(common.ERR_DEVICE_SUBSCRIBED).Status(http.StatusConflict).Log().Payload(nil).Write(w)
			return
		}
	}

	// append new device into the devices array for such user
	devs = append(devs, payload)

	if saved := db.SetOne(db.SubscriptionCache, callerID, devs); !saved {
		l.Msg(common.ERR_SUBSCRIPTION_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, device subscription added").Status(http.StatusCreated).Log().Payload(nil).Write(w)
	return
}

// sendNotification is the push pkg handler function for sending new notification(s).
//
// @Summary      Send a notification
// @Description  Send a notification
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Success      201  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Router       /push/notification/{postID} [post]
func sendNotification(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	// this user ID points to the replier
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	payload := struct {
		OriginalID string `json:"original_post"`
	}{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// fetch related data from cachces
	post, ok := db.GetOne(db.FlowCache, payload.OriginalID, models.Post{})
	if !ok {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	devs, ok := db.GetOne(db.SubscriptionCache, post.Nickname, []models.Device{})
	if !ok {
		l.Msg(common.ERR_SUBSCRIPTION_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// do not notify the same person --- OK condition
	if post.Nickname == callerID {
		l.Msg(common.ERR_PUSH_SELF_NOTIF).Status(http.StatusOK).Log().Payload(nil).Write(w)
		return
	}

	// do not notify user --- notifications disabled --- OK condition
	if len(devs) == 0 {
		l.Msg(common.ERR_PUSH_DISABLED_NOTIF).Status(http.StatusOK).Log().Payload(nil).Write(w)
		return
	}

	// compose the body of this notification
	body, _ := json.Marshal(app.Notification{
		Title: "littr reply",
		Icon:  "/web/apple-touch-icon.png",
		Body:  callerID + " replied to your post",
		Path:  "/flow/post/" + post.ID,
	})

	// fire notification goroutines
	SendNotificationToDevices(post.Nickname, devs, body, l)

	l.Msg("ok, notification(s) are being sent").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// deleteSubscription is the push pkg handler function to ensure a given subscription deleted from the database.
//
// @Summary      Delete a subscription
// @Description  delete a subscription
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription/{uuid} [delete]
func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	payload := struct {
		UUID string `json:"device_uuid"`
	}{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	uuid := payload.UUID
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	devices, ok := db.GetOne(db.SubscriptionCache, callerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_SUBSCRIPTION_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var newDevices []models.Device

	for _, dev := range devices {
		if dev.UUID == uuid {
			// do not include this device anymore
			continue
		}

		if dev.UUID == "" {
			// do not include blank-labeled devices
			continue
		}

		newDevices = append(newDevices, dev)
	}

	if saved := db.SetOne(db.SubscriptionCache, callerID, newDevices); !saved {
		l.Msg(common.ERR_SUBSCRIPTION_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, subscription deleted").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}
