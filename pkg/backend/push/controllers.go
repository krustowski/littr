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
// @Success      200  {object}   common.Response
// @Failure 	 401  {object}   common.Response
// @Router       /push/vapid [get]
func fetchVAPIDKey(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "push")

	caller, _ := r.Context().Value("nickname").(string)
	if caller == "" {
		resp.Message = "client unauthorized"
		resp.Code = http.StatusUnauthorized

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, sending VAPID public key"
	resp.Key = os.Getenv("VAPID_PUB_KEY")
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// updateSubscription is the push pkg's handler function used to update an existing subscription.
//
// @Summary      Update the notification subscription tag
// @Description  Update the notification subscription tag
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /push/subscription/{uuid}/mention [put]
// @Router       /push/subscription/{uuid}/reply [put]
func updateSubscription(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "push")

	path := r.URL.Path
	caller, _ := r.Context().Value("nickname").(string)
	payload := models.Device{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache later on
	devs, _ := db.GetOne(db.SubscriptionCache, caller, []models.Device{})

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

	if saved := db.SetOne(db.SubscriptionCache, caller, devs); !saved {
		resp.Code = http.StatusInternalServerError
		resp.Message = "cannot save new subscription"

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, device subscription updated"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// subscribeToNotifications is the push pkg's handler function to ensure sent device has subscribed to notifications.
//
// @Summary      Add the notification subscription
// @Description  add the notification subscription
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      409  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /push/subscription [post]
func subscribeToNotifications(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "push")

	caller, _ := r.Context().Value("nickname").(string)
	payload := models.Device{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache
	devs, _ := db.GetOne(db.SubscriptionCache, caller, []models.Device{})

	for _, dev := range devs {
		if dev.UUID == payload.UUID {
			// we have just found a match, thus request was fired twice, or someone tickles the API
			resp.Message = "backend notice: this device has already been registered for a subscription"
			resp.Code = http.StatusConflict

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}
	}

	// append new device into the devices array for such user
	devs = append(devs, payload)

	if saved := db.SetOne(db.SubscriptionCache, caller, devs); !saved {
		resp.Code = http.StatusInternalServerError
		resp.Message = "cannot save new subscription"

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, device subscription added"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// sendNotification is the push pkg handler function for sending new notification(s).
//
// @Summary      Send a notification
// @Description  Send a notification
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Success      201  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Router       /push/notification/{postID} [post]
func sendNotification(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "push")

	// this user ID points to the replier
	caller, _ := r.Context().Value("nickname").(string)

	// hm, this looks kinda sketchy...
	// TODO: make this more readable
	original := struct {
		ID string `json:"original_post"`
	}{}

	if err := common.UnmarshalRequestData(r, &original); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// fetch related data from cachces
	post, _ := db.GetOne(db.FlowCache, original.ID, models.Post{})
	devs, _ := db.GetOne(db.SubscriptionCache, post.Nickname, []models.Device{})
	//user, _ := db.GetOne(db.UserCache, post.Nickname, users.User{})

	// do not notify the same person --- OK condition
	if post.Nickname == caller {
		resp.Message = "do not send notifs to oneself"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// do not notify user --- notifications disabled --- OK condition
	if len(devs) == 0 {
		resp.Message = "notifications disabled for such user"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// compose the body of this notification
	body, _ := json.Marshal(app.Notification{
		Title: "littr reply",
		Icon:  "/web/apple-touch-icon.png",
		Body:  caller + " replied to your post",
		Path:  "/flow/post/" + post.ID,
	})

	SendNotificationToDevices(post.Nickname, devs, body, l)

	resp.Message = "ok, notification(s) sent"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// deleteSubscription is the push pkg handler function to ensure a given subscription deleted from the database.
//
// @Summary      Delete a subscription
// @Description  delete a subscription
// @Tags         push
// @Accept       json
// @Produce      json
// @Success      200  {object}   common.Response
// @Failure      400  {object}   common.Response
// @Failure      500  {object}   common.Response
// @Router       /push/subscription/{uuid} [delete]
func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, "push")

	// this user ID points to the replier
	caller, _ := r.Context().Value("nickname").(string)

	payload := struct {
		UUID string `json:"device_uuid"`
	}{}

	if err := common.UnmarshalRequestData(r, &payload); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	uuid := payload.UUID
	if uuid == "" {
		resp.Message = "backend error: blank UUID received!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	devices, _ := db.GetOne(db.SubscriptionCache, caller, []models.Device{})

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

	if saved := db.SetOne(db.SubscriptionCache, caller, newDevices); !saved {
		resp.Message = "new subscription state of devices could not be saved"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, subscription deleted"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
	return
}
