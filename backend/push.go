package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func subscribeToNotifs(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "push")

	var sub webpush.Subscription

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &sub); err != nil {
		resp.Message = "backend error: cannot unmarshall request data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	caller, _ := r.Context().Value("nickname").(string)

	// fetch existing (or blank) subscription array for such caller, and add new sub.
	subs, _ := getOne(SubscriptionCache, caller, []webpush.Subscription{})
	subs = append(subs, sub)

	if saved := setOne(SubscriptionCache, caller, subs); !saved {
		resp.Code = http.StatusInternalServerError
		resp.Message = "cannot save new subscription"

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, subscription added"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

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

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	original := struct {
		ID string `json:"original_post"`
	}{}

	if err := json.Unmarshal(data, &original); err != nil {
		resp.Message = "backend error: cannot unmarshall request data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// fetch related data from cachces
	post, _ := getOne(FlowCache, original.ID, models.Post{})
	subs, _ := getOne(SubscriptionCache, post.Nickname, []webpush.Subscription{})
	user, _ := getOne(UserCache, post.Nickname, models.User{})

	// do not notify the same person --- OK condition
	if post.Nickname == caller {
		resp.Message = "do not send notifs to oneself"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// do not notify user --- notifications disabled --- OK condition
	if &subs == nil {
		resp.Message = "notifications disabled for such user"
		resp.Code = http.StatusOK

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	for _, sub := range subs {
		// prepare and send new notification
		go func(sub webpush.Subscription) {
			body, _ := json.Marshal(app.Notification{
				Title: "littr reply",
				Icon:  "/web/apple-touch-icon.png",
				Body:  caller + " replied to your post",
				Path:  "/flow/post/" + post.ID,
			})

			// fire a notification
			res, err := webpush.SendNotification(body, &sub, &webpush.Options{
				VAPIDPublicKey:  user.VapidPubKey,
				VAPIDPrivateKey: user.VapidPrivKey,
				TTL:             300,
			})
			if err != nil {
				resp.Code = http.StatusInternalServerError
				resp.Message = "cannot send a notification: " + err.Error()

				l.Println(resp.Message, resp.Code)
				resp.Write(w)
				return
			}

			defer res.Body.Close()
		}(sub)
	}

	resp.Message = "ok, notification fired"
	resp.Code = http.StatusCreated

	l.Println(resp.Message, resp.Code)

	resp.Write(w)
}
