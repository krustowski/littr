package common

import (
	//"context"
	"fmt"
	"net/http"
	"net/url"
	//"os"
	//"os/signal"
	//"syscall"
	"time"

	//"go.vxn.dev/littr/pkg/config"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/tmaxmax/go-sse"
)

// Default response validator.
// https://pkg.go.dev/github.com/tmaxmax/go-sse@v0.8.0#ResponseValidator
var DefaultValidator sse.ResponseValidator = func(r *http.Response) error {
	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("expected status code %d %s, received %d %s", http.StatusOK, http.StatusText(http.StatusOK), r.StatusCode, http.StatusText(r.StatusCode))
	}
	cts := r.Header.Get("Content-Type")
	//ct := contentType(cts)
	if expected := "text/event-stream"; cts != expected {
		return fmt.Errorf("expected content type to have %q, received %q", expected, cts)
	}
	return nil
}

// Noop response validator.
// https://pkg.go.dev/github.com/tmaxmax/go-sse@v0.8.0#ResponseValidator
var NoopValidator sse.ResponseValidator = func(_ *http.Response) error {
	return nil
}

func getRequestURL(sub string) string {
	q := url.Values{}
	switch sub {
	case "all":
		q.Add("topic", "numbers")
		q.Add("topic", "metrics")
	case "numbers", "metrics":
		q.Set("topic", sub)
	default:
		panic(fmt.Errorf("unexpected subscription topic %q", sub))
	}

	return URL + "/api/v1/live?" + q.Encode()
}

// URL is a simple lambda function to retrieve the URL for a new SSE connection.
var URL = func() string {
	// Use APP_URL_MAIN env variables in prod and staging environments.
	if app.Getenv("APP_URL_MAIN") != "" {
		return "https://" + app.Getenv("APP_URL_MAIN")
	}

	// Local development use only.
	return "http://localhost:8080"
}()

// Custom HTTP client.
var Client = sse.Client{
	// Standard HTTP client.
	HTTPClient: &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			// Idle = keeplive conn
			// https://pkg.go.dev/net/http#Transport
			MaxIdleConns:       1,
			IdleConnTimeout:    20 * time.Second,
			DisableCompression: true,
			DisableKeepAlives:  false,
		},
	},
	// Callback function when the connection is to be reastablished.
	OnRetry: func(err error, duration time.Duration) {
		fmt.Printf("conn error: %v\n", err)
		time.Sleep(duration)
	},
	// Validation of the response content-type mainly, e.g. DefaultValidator.
	ResponseValidator: DefaultValidator,
	// The connection tuning.
	Backoff: sse.Backoff{
		InitialInterval: 500 * time.Millisecond,
		Multiplier:      float64(1.5),
		// Jitter: range (0, 1)
		Jitter:         float64(0.5),
		MaxInterval:    2 * time.Second,
		MaxElapsedTime: 2 * time.Second,
		MaxRetries:     10,
	},
}

/*func sSEClient() {
	// Prepare the context for the client shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sub := "all"

	// Compose a new connection with the context.
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, getRequestURL(sub), http.NoBody)
	_ = sse.NewConnection(req)
}*/

// Subscribe to any event, regardless the type.
/*conn.SubscribeToAll(func(event sse.Event) {
ctx.NewActionWithValue("generic-event", event)

// Print all events.
fmt.Printf("%s: %s\n", event.Type, event.Data)

/*switch event.Type {
case "keepalive", "ops":
	fmt.Printf("%s: %s\n", event.Type, event.Data)
case "server-stop":
	fmt.Println("server closed!")
	h.sseCancel()
default: // no event name*/
//}
//})

//
//
//
