// Package to provide a server-side instance for SSE.
package live

import (
	"time"

	chi "github.com/go-chi/chi/v5"
	sse "github.com/tmaxmax/go-sse"

	// Those two should be abaddoned already.
	//sse "github.com/alexandrevicenzi/go-sse"
	//sse "github.com/r3labs/sse/v2"

	"go.vxn.dev/littr/pkg/config"
)

var Streamer *sse.Server

// Get live flow event stream
//
// @Summary      Get real-time posts event stream (SSE stream)
// @Description  get real-time posts event stream
// @Tags         live
// @Produce      text/event-stream
// @Success      200  {object}  string
// @Failure      500  {object}  nil
// @Router       /live [get]
func beat() {
	for {
		// Break the loop if Streamer is nil.
		/*if Streamer == nil {
			break
		}

		BroadcastMessage("heartbeat", "keepalive")
		time.Sleep(time.Second * config.HEARTBEAT_SLEEP_TIME)*/

		// New implementation (other lib).
		msg := &sse.Message{}
		id, err := sse.NewID("keepalive")
		if err != nil {
			// ???
			break
		}

		msg.ID = id

		msg.AppendData("heartbeat")

		if Streamer != nil {
			_ = Streamer.Publish(msg)
		}

		// Sleep for the given period of time.
		time.Sleep(time.Second * config.HEARTBEAT_SLEEP_TIME)
	}
}

// ID, data, event
// https://github.com/alexandrevicenzi/go-sse/blob/master/message.go#L23

// BroadcastMessage is a wrapper function for a SSE message sending.
func BroadcastMessage(data, eventName string) {
	if data == "" {
		return
	}

	// Compose a message.
	msg := &sse.Message{}
	msg.AppendData(data)

	if eventName == "" {
		eventName = "message"
	}

	// Ensure a proper ID is used.
	id, err := sse.NewID(eventName)
	if err != nil {
		return
	}
	msg.ID = id

	if Streamer != nil {
		_ = Streamer.Publish(msg)

		/*Streamer.SendMessage("/api/v1/live",
		sse.NewMessage("", data, eventName))*/
	}
	return
}

func Router() chi.Router {
	r := chi.NewRouter()

	// Core SSE server struct initialization.
	Streamer = &sse.Server{
		// Joe is the default provider.
		Provider: &sse.Joe{
			// Replays only valid events, that expire after 5 minutes.
			ReplayProvider: &sse.ValidReplayProvider{TTL: time.Minute * 5},
		},
		//OnSession:
		//Logger:
	}

	// Run the keepalive pacemaker.
	go beat()

	r.Mount("/", Streamer)

	return r
}
