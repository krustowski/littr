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

	chi "github.com/go-chi/chi/v5"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// fetchVAPIDKey is a helper function to fetch new server VAPID key pair.
//
// @Summary      Get a VAPID key pair
// @Description  get a VAPID key pair
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.APIResponse{data=push.getchVAPIDKey.responseData}
// @Failure 	 400  {object}   common.APIResponse
// @Failure 	 401  {object}   common.APIResponse{data=auth.authHandler.responseData}
// @Router       /push/vapid [get]
func fetchVAPIDKey(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	type responseData struct {
		Key string `json:"key"`
	}

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// prepare the payload
	pl := &responseData{
		Key: os.Getenv("VAPID_PUB_KEY"),
	}

	l.Msg("ok, sending public VAPID key").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// updateSubscription is the push pkg's handler function used to update an existing subscription.
//
// @Summary      Update the notification subscription tag
// @Description  Update the notification subscription tag
// @Tags         push
// @Accept       json
// @Produce      json
// @Param        uuid path string true "UUID of a device to update"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription/{uuid}/mention [put]
// @Router       /push/subscription/{uuid}/reply [put]
func updateSubscription(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache later on
	devs, ok := db.GetOne(db.SubscriptionCache, l.CallerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_DEVICE_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	var device models.Device

	// get the specified device by its UUID
	for _, dev := range devs {
		if dev.UUID == uuid {
			device = dev
		}
	}

	// "switch" the subscription type
	path := r.URL.Path

	var tag string

	// decide which tag we are processing at the moment
	if strings.Contains(path, "mention") {
		tag = "mention"
	} else if strings.Contains(path, "reply") {
		tag = "reply"
	}

	// append the tag if not present already
	if !helpers.Contains(device.Tags, tag) {
		device.Tags = append(device.Tags, tag)
	} else {
		// remove the existing tag
		tags := device.Tags

		var newTags []string

		// loop over tags and yeet the one existing and requested
		for _, t := range tags {
			if t == tag {
				continue
			}
			newTags = append(newTags, t)
		}

		// update device's tags
		device.Tags = newTags
	}

	// loop through loaded devices and look for the requested one
	// if found, update the existing requested device
	found := false
	for idx, dev := range devs {
		// requested dev's UUID matches loaded device's UUID
		if dev.UUID == device.UUID {
			// save the requested device to the device array
			devs[idx] = device
			found = true
		}
	}

	// append new device into the devices array for such user, if not found before
	if !found {
		devs = append(devs, device)
	}

	// save the updated device array back to the database
	if saved := db.SetOne(db.SubscriptionCache, l.CallerID, devs); !saved {
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
// @Param    	 request body models.Device true "device to subscribe"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      409  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription [post]
func subscribeToNotifications(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var device models.Device

	// decode the incoming request's body
	if err := common.UnmarshalRequestData(r, &device); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache
	devs, ok := db.GetOne(db.SubscriptionCache, l.CallerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_DEVICE_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// loop through the devices and check its presence (present = already subscribed)
	for _, dev := range devs {
		if dev.UUID == device.UUID {
			// we have just found a match, thus request was fired twice, or someone tickles the API
			l.Msg(common.ERR_DEVICE_SUBSCRIBED_ALREADY).Status(http.StatusConflict).Log().Payload(nil).Write(w)
			return
		}
	}

	// append new device into the devices array for such user
	devs = append(devs, device)

	// save the device array back to the database
	if saved := db.SetOne(db.SubscriptionCache, l.CallerID, devs); !saved {
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
// @Produce      json
// @Param        postID path string true "original post's ID"
// @Success      200  {object}   common.APIResponse
// @Success      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Router       /push/notification/{postID} [post]
func sendNotification(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	postID := chi.URLParam(r, "postID")
	if postID == "" {
		l.Msg(common.ERR_POSTID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch related data from the database
	post, ok := db.GetOne(db.FlowCache, postID, models.Post{})
	if !ok {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// fetch notification receiver's device list
	devs, ok := db.GetOne(db.SubscriptionCache, post.Nickname, []models.Device{})
	if !ok {
		// it is OK for this to return nothing, loop can handle it later...
		//l.Msg(common.ERR_SUBSCRIPTION_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		//return
	}

	// do not notify the same person --- OK condition
	if post.Nickname == l.CallerID {
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
		Body:  l.CallerID + " replied to your post",
		Path:  "/flow/post/" + post.ID,
	})

	opts := &NotificationOpts{
		Receiver: post.Nickname,
		Devices:  &devs,
		Body:     &body,
		Logger:   l,
	}

	// send the webpush notification(s)
	SendNotificationToDevices(opts)

	l.Msg("ok, notification(s) are being sent").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// deleteSubscription is the push pkg handler function to ensure a given subscription deleted from the database.
//
// @Summary      Delete a subscription
// @Description  delete a subscription
// @Tags         push
// @Produce      json
// @Param        uuid path string true "UUID of a device to delete"
// @Success      200  {object}   common.APIResponse
// @Failure      400  {object}   common.APIResponse
// @Failure      404  {object}   common.APIResponse
// @Failure      500  {object}   common.APIResponse
// @Router       /push/subscription/{uuid} [delete]
func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, "push")

	// skip blank callerID
	if l.CallerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	uuid := chi.URLParam(r, "uuid")
	if uuid == "" {
		l.Msg(common.ERR_PUSH_UUID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache later on
	devs, ok := db.GetOne(db.SubscriptionCache, l.CallerID, []models.Device{})
	if !ok {
		l.Msg(common.ERR_DEVICE_NOT_FOUND).Status(http.StatusNotFound).Log().Payload(nil).Write(w)
		return
	}

	// prepare a new array for remaining devices
	var newDevices []models.Device

	// loop through fetched devices, skip the requested one as well as the blank ones
	for _, dev := range devs {
		if dev.UUID == uuid {
			// do not include this device anymore
			continue
		}

		//if dev.UUID == "" || dev == (models.Device{}) {
		if dev.UUID == "" {
			// do not include blank-labeled devices or blank ones in general
			continue
		}

		// append non matching device
		newDevices = append(newDevices, dev)
	}

	// save the update device list
	if saved := db.SetOne(db.SubscriptionCache, l.CallerID, newDevices); !saved {
		l.Msg(common.ERR_SUBSCRIPTION_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, subscription deleted").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}
