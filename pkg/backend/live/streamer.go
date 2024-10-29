package live

import (
	"net/http"
	"time"

	sse "github.com/tmaxmax/go-sse"

	// Those two should be abaddoned already.
	//sse "github.com/alexandrevicenzi/go-sse"
	//sse "github.com/r3labs/sse/v2"

	"go.vxn.dev/littr/pkg/config"
)

const (
	topicMetrics       = "metrics"
	topicRandomNumbers = "numbers"
)

// Core SSE server struct initialization, the server implements http.Handler interface.
var Streamer = &sse.Server{
	// Joe is the default pubsub service provider.
	Provider: &sse.Joe{
		// Replays only valid events, that expire after 5 minutes.
		ReplayProvider: &sse.ValidReplayProvider{
			TTL:        time.Minute * 5,
			GCInterval: time.Minute,
			AutoIDs:    true,
		},
	},
	// Custom callback function when a SSE session is started.
	OnSession: func(s *sse.Session) (sse.Subscription, bool) {
		// Fetch the topic list.
		topics := s.Req.URL.Query()["topic"]

		// Loop over the topic list to determine the returned Subscription structure.
		for _, topic := range topics {
			// The topic is unknown or invalid.
			if topic != topicRandomNumbers && topic != topicMetrics {
				// Do not send a pre-superfluous response.
				//fmt.Fprintf(s.Res, "invalid topic %q; supported are %q, %q", topic, topicRandomNumbers, topicMetrics)
				//s.Res.WriteHeader(http.StatusBadRequest)
				//return sse.Subscription{}, false
			}
		}

		if len(topics) == 0 {
			// Provide default topics, if none are given.
			topics = []string{topicRandomNumbers, topicMetrics}
		}

		// Return a new SSE subscription.
		return sse.Subscription{
			Client:      s,
			LastEventID: s.LastEventID,
			Topics:      append(topics, sse.DefaultTopic), // The shutdown message is sent on the default topic.
		}, true
	},
	//Logger:
}

// The very keepalive pacemaker.
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
		if Streamer == nil {
			break
		}

		// Send the message.
		BroadcastMessage("heartbeat", "keepalive")

		// Sleep for the given period of time.
		time.Sleep(time.Second * config.HEARTBEAT_SLEEP_TIME)
	}
}

// BroadcastMessage is a wrapper function for a SSE message sending.
func BroadcastMessage(data, eventName string) {
	// Exit if Streamer is nil.
	if Streamer == nil {
		return
	}

	// Refuse empty data.
	if data == "" {
		return
	}

	// Compose a message.
	msg := &sse.Message{}

	// Ensure a default event ID set.
	if eventName == "" {
		eventName = "message"
	}

	// Ensure a valid ID is used.
	id, err := sse.NewID(eventName)
	if err != nil {
		return
	}
	msg.ID = id

	// Ensure a valid event Type is used.
	typ, err := sse.NewType(eventName)
	if err != nil {
		return
	}
	msg.Type = typ

	// Append any given data to the event.
	msg.AppendData(data)

	// Broadcast the message to the subscribers.
	if Streamer != nil {
		_ = Streamer.Publish(msg)
	}
	return
}

// cors is a simple HTTP Handler wrapper that serves the CORS-ready headers.
// https://github.com/tmaxmax/go-sse/blob/master/cmd/complex/main.go
func cors(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://dev.littr.eu:8080")
		h.ServeHTTP(w, r)
	})
}
