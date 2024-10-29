package common

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	//"go.vxn.dev/littr/pkg/config"

	"github.com/tmaxmax/go-sse"
)

var URL = func() string {
	// Use APP_URL_MAIN env variables in prod and staging environments.
	if os.Getenv("APP_URL_MAIN") != "" {
		return "https://" + os.Getenv("APP_URL_MAIN")
	}

	// Local development use only.
	return "http://localhost:8080"
}()

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

func SSEClient() {
	// Prepare the context for the client shutdown.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	sub := "all"

	// Compose a new connection with the context.
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, getRequestURL(sub), http.NoBody)
	_ = sse.NewConnection(req)
}

// Custom HTTP client.
var Client = sse.Client{
	// Standard HTTP client.
	HTTPClient: &http.Client{
		Timeout: 30 * time.Second,
	},
	// Callback function when the connection is to be reastablished.
	OnRetry: func(err error, duration time.Duration) {
		fmt.Printf("conn error: %v\n", err)
		time.Sleep(duration)
	},
	// Validation of the response content-type mainly.
	ResponseValidator: DefaultValidator,
	// The connection tuning.
	Backoff: sse.Backoff{
		InitialInterval: 1000 * time.Millisecond,
		Multiplier:      float64(1.5),
		// Jitter: range (0, 1)
		Jitter:         float64(0.5),
		MaxInterval:    10 * time.Second,
		MaxElapsedTime: 45 * time.Second,
		MaxRetries:     15,
	},
}
