package models

import (
	"time"

	"github.com/SherClockHolmes/webpush-go"
)

// Helper struct to see how the data is stored in the database.
type Devices struct {
	Devices map[string]Device `json:"items"`
}

// SubscriptionDevice
type Device struct {
	// Unique identification of the app on the current device.
	// https://go-app.dev/reference#Context
	UUID         string               `json:"uuid"`

	// Timestamp of the subscription creation.
	TimeCreated  time.Time            `json:"time_created"`

	// The very subscription struct/details.
	Subscription webpush.Subscription `json:"subscription"`
}
