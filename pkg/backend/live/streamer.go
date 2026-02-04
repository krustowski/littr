package live

import (
	"net/http"
	"time"

	sse "github.com/tmaxmax/go-sse"

	// Those two should be abaddoned already.
	//sse "github.com/alexandrevicenzi/go-sse"
	//sse "github.com/r3labs/sse/v2"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/config"
)

const (
	topicMetrics       = "metrics"
	topicRandomNumbers = "numbers"
)

var replayer = func() *sse.ValidReplayer {
	rep, err := sse.NewValidReplayer(time.Minute*4, true)
	if err != nil {
		return nil
	}

	return rep
}()

// Core SSE server as the HTTP handler wrapper.
var Streamer = &sse.Server{
	// Joe is the default pubsub service provider.
	Provider: &sse.Joe{
		// Replays only valid events, that expire after 5 minutes.
		Replayer: replayer,
		/*ReplayProvider: &sse.ValidReplayProvider{
			TTL:        time.Minute * 5,
			GCInterval: time.Minute,
			AutoIDs:    true,
		},*/
	},
	// Custom callback function when a SSE session is started.
	OnSession: func(w http.ResponseWriter, r *http.Request) (topics []string, allowed bool) {
		topics = r.URL.Query()["topic"]

		// The shutdown message is sent via the default topic
		topics = append(topics, sse.DefaultTopic)

		// TODO: secure this
		allowed = true

		return
	},
	//Logger:
}

// The very keepalive pacemaker.
//
//	@Summary		Get real-time server-sent event stream (SSE stream)
//	@Description		Calling this endpoint creates a SSE subscription to receive the server-sent event stream. The connection type is set to keep-alive, so the common request will appear as "timing-out".
//	@Tags			live
//	@Produce		text/event-stream
//	@Success		200	{object} 	string		"The connection success. Typically appears when the stream ends gracefully."
//	@Failure		500	{object}	nil		"A generic network problem when connecting to the stream."
//	@Router			/live [get]
func beat() {
	for {
		// Break the loop if Streamer is nil.
		if Streamer == nil {
			l := common.NewLogger(nil, "pacemaker")

			l.Msg("the SSE streamer is nil, stopping the pacemaker...").Status(http.StatusInternalServerError).Log()
			break
		}

		// Send the message.
		BroadcastMessage(EventPayload{Data: "heartbeat", Type: "keepalive"})

		// Sleep for the given period of time.
		time.Sleep(time.Second * config.StreamerHeartbeatPeriod)
	}
}

// EventPayload is the metastructure to organize the SSE event's data association better.
// It is an input for the BroadcastMessage function.
type EventPayload struct {
	ID   string
	Type string
	Data string
}

// BroadcastMessage is a wrapper function for a SSE message sending.
func BroadcastMessage(payload EventPayload) {
	// Exit if Streamer is nil.
	if Streamer == nil {
		return
	}

	// Refuse empty data.
	if payload.Data == "" {
		return
	}

	// Compose a message.
	msg := &sse.Message{}

	// Ensure a default event ID set.
	if payload.Type == "" {
		payload.Type = "message"
	}

	// Ensure a valid ID is used.
	id, err := sse.NewID(payload.ID)
	if err != nil {
		return
	}
	msg.ID = id

	// Ensure a valid event Type is used.
	typ, err := sse.NewType(payload.Type)
	if err != nil {
		return
	}
	msg.Type = typ

	// Append any given data to the event.
	msg.AppendData(payload.Data)

	// Broadcast the message to the subscribers.
	if Streamer != nil {
		_ = Streamer.Publish(msg)
	}
}
