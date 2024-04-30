package push

import (
	"io"
	"net/http"
	"os"
	"strings"

	"go.savla.dev/littr/pkg/backend/common"
	"go.savla.dev/littr/pkg/helpers"
	"go.savla.dev/littr/pkg/models"

	"github.com/SherClockHolmes/webpush-go"
)

func SendNotificationToDevices(devs []models.Device, body []byte, l *common.Logger) {
	tag := ""
	if strings.Contains(string(body), "reply") {
		tag = "reply"
	} else if strings.Contains(string(body), "mention") {
		tag = "mention"
	}

	// range devices
	for _, dev := range devs {
		if dev.UUID == "" {
			continue
		}

		// skip devices unsubscribed to such notification tag
		if !helpers.Contains(dev.Tags, tag) {
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
