package models

import (
	"time"

	"github.com/SherClockHolmes/webpush-go"
	//"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// Helper struct to see how the data is stored in the database.
type Devices struct {
	//Devices map[string]Device `json:"items"`
	Devices []Device `json:"items"`
}

// SubscriptionDevice
type Device struct {
	// Unique identification of the app on the current device.
	// https://go-app.dev/reference#Context
	UUID string `json:"uuid"`

	// Timestamp of the subscription creation.
	TimeCreated time.Time `json:"time_created"`

	// Timestamp of the last notification sent through this device.
	TimeLastUsed time.Time `json:"time_last_used"`

	// List of labels for such device.
	Tags []string `json:"tags,omitempty"`

	// The very subscription struct/details.
	//Subscription app.NotificationSubscription `json:"subscription"`
	Subscription webpush.Subscription `json:"subscription"`
}
