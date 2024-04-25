package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"go.savla.dev/littr/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// GET /api/push/vapid
func fetchVAPIDKey(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "push")

	caller, _ := r.Context().Value("nickname").(string)
	if caller == "" {
		resp.Message = "client unauthorized "
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

// POST /api/push
func subscribeToNotifs(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "push")

	caller, _ := r.Context().Value("nickname").(string)
	payload := models.Device{}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &payload); err != nil {
		resp.Message = "backend error: cannot unmarshall request data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// let us check this device
	// we are about to loop through []models.Device fetched from SubscriptionCache
	devs, _ := getOne(SubscriptionCache, caller, []models.Device{})

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

	if saved := setOne(SubscriptionCache, caller, devs); !saved {
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

// PUT /api/push
func sendNotif(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "push")

	// this user ID points to the replier
	caller, _ := r.Context().Value("nickname").(string)

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// hm, this looks kinda sketchy...
	// TODO: make this more readable
	original := struct {
		ID string `json:"original_post"`
	}{}

	if err := json.Unmarshal(reqBody, &original); err != nil {
		resp.Message = "backend error: cannot unmarshall request data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// fetch related data from cachces
	post, _ := getOne(FlowCache, original.ID, models.Post{})
	devs, _ := getOne(SubscriptionCache, post.Nickname, []models.Device{})
	//user, _ := getOne(UserCache, post.Nickname, models.User{})

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

	sendNotificationToDevices(devs, body, l)

	resp.Message = "ok, notification(s) fired"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

func sendNotificationToDevices(devs []models.Device, body []byte, l *Logger) {
	// range devices
	for _, dev := range devs {
		if dev.UUID == "" {
			continue
		}

		// run this async not to make client wait too much
		//
		// IMPORTANT NOTE: do not write headers in the goroutine --- this will crash the server on nil pointer dereference
		// and memory segment violation
		go func(dev models.Device, body []byte) {
			sub := dev.Subscription

			// fire a notification
			res, err := webpush.SendNotification(body, &sub, &webpush.Options{
				Subscriber:      os.Getenv("VAPID_SUBSCRIBER"),
				VAPIDPublicKey:  os.Getenv("VAPID_PUB_KEY"),
				VAPIDPrivateKey: os.Getenv("VAPID_PRIV_KEY"),
				TTL:             30,
				Urgency:         webpush.UrgencyNormal,
			})
			if err != nil {
				code := http.StatusInternalServerError
				message := "cannot send a notification: " + err.Error()

				l.Println(message, code)
				return
			}

			defer res.Body.Close()

			bodyBytes, err := io.ReadAll(res.Body)
			if err != nil {
				// TODO: handle this
				//log.Fatal(err)
			}

			bodyString := string(bodyBytes)
			if bodyString == "" {
				bodyString = "(blank)"
			}

			code := res.StatusCode
			message := "push gorutine: response from the counterpart: " + bodyString
			l.Println(message, code)
			return
		}(dev, body)
	}
}

func deleteSubscription(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "push")

	// this user ID points to the replier
	caller, _ := r.Context().Value("nickname").(string)

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	payload := struct {
		UUID string `json:"device_uuid"`
	}{}

	if err := json.Unmarshal(reqBody, &payload); err != nil {
		resp.Message = "backend error: cannot unmarshall request data: " + err.Error()
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

	devices, _ := getOne(SubscriptionCache, caller, []models.Device{})

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

	if saved := setOne(SubscriptionCache, caller, newDevices); !saved {
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
