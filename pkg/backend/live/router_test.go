package live

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"go.vxn.dev/littr/pkg/config"

	chi "github.com/go-chi/chi/v5"
	sse "github.com/tmaxmax/go-sse"
)

var streamerTestURI = "/api/v1/live"

func TestLiveRouterWithStreamer(t *testing.T) {
	r := chi.NewRouter()

	// For the Streamer configuration check pkg/backend/live/streamer.go
	r.Mount(streamerTestURI, Streamer)

	// Fetch test net listener and test HTTP server configuration.
	listener := config.PrepareTestListenerWithPort(t, config.DEFAULT_TEST_SSE_PORT)
	defer listener.Close()

	ts := config.PrepareTestServer(t, listener, r)
	ts.Start()
	defer ts.Close()

	// Run the testing keepalive pacemaker.
	go testBeat()

	var wg sync.WaitGroup

	// Spin-off a client SSE goroutine and wait till it dead.
	wg.Add(1)
	go testConnectorSSE(t, &wg, "http://localhost:"+config.DEFAULT_TEST_SSE_PORT+streamerTestURI)

	var timeout = 4 * time.Second

	// Spin another goroutine to handle the deadline.
	go func() {
		time.Sleep(timeout)

		// Fetch a context to send to gracefully shutdown the HTTP server.
		sctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		BroadcastMessage(EventPayload{Data: "server-stop", Type: "close"})

		// Terminate the SSE server.
		Streamer.Shutdown(sctx)
	}()

	// Wait for the client to exit.
	wg.Wait()

	return
}

func testConnectorSSE(t *testing.T, wg *sync.WaitGroup, endpoint string) {
	if t == nil || wg == nil {
		return
	}

	var eventReceived bool = false

	defer func() {
		if !eventReceived {
			t.Errorf("client is closing but no event has been received")
		}

		wg.Done()
	}()

	// Create a cancellable context for the HTTP request.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Prepare a HTTP request
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	conn := sse.NewConnection(r)

	// Callback function called when any event is received.
	conn.SubscribeToAll(func(event sse.Event) {
		if event.Type != "keepalive" && event.Data != "heartbeat" {
			t.Errorf("non-heartbeat event received")
			t.Errorf("%s: %s\n", event.Type, event.Data)
		}
		eventReceived = true
		cancel()
	})

	// Make a connection to the SSE streamer, wait for errors or context cancel.
	if err := conn.Connect(); err != nil && err.Error() != "context canceled" {
		t.Errorf(err.Error())
	}

	return
}

func testBeat() {
	for {
		// Break the loop if Streamer is nil.
		if Streamer == nil {
			break
		}

		// Sleep for the given period of time.
		time.Sleep(time.Millisecond * 2500)

		// Send the message.
		BroadcastMessage(EventPayload{Data: "heartbeat", Type: "keepalive"})
	}
}
