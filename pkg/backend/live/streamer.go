package live

import (
	"time"

	chi "github.com/go-chi/chi/v5"
	//sse "github.com/tmaxmax/go-sse"
	sse "github.com/alexandrevicenzi/go-sse"
	//sse "github.com/r3labs/sse/v2"

	cfg "go.vxn.dev/littr/configs"
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
	// ID, data, event
	// https://github.com/alexandrevicenzi/go-sse/blob/master/message.go#L23
	for {
		if Streamer == nil {
			break
		}

		BroadcastMessage("heartbeat", "keepalive")
		time.Sleep(time.Second * cfg.HEARTBEAT_SLEEP_TIME)
	}

	return
}

// wrapper function for SSE messages sending
func BroadcastMessage(data, eventName string) {
	if data == "" {
		return
	}

	if eventName == "" {
		eventName = "message"
	}

	if Streamer != nil {
		Streamer.SendMessage("/api/v1/live",
			sse.NewMessage("", data, eventName))
	}

	return
}

func Router() chi.Router {
	r := chi.NewRouter()

	Streamer = sse.NewServer(&sse.Options{
		Logger: nil,
	})

	go beat()

	r.Mount("/", Streamer)

	return r
}
