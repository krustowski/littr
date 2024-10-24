package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"go.vxn.dev/littr/pkg/config"

	"github.com/tmaxmax/go-sse"
)

var URL = func() string {
	if os.Getenv("SSE_CLIENT_URL") != "" {
		return os.Getenv("SSE_CLIENT_URL")
	}

	return "http://localhost:" + config.ServerPort
}()

func main() {
	var sub string
	flag.StringVar(&sub, "sub", "all", "The topics to subscribe to. Valid values are: all, numbers, metrics")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, getRequestURL(sub), http.NoBody)
	conn := sse.NewConnection(r)

	conn.SubscribeToAll(func(event sse.Event) {
		switch event.Type {
		case "keepalive", "ops":
			fmt.Printf("%s: %s\n", event.Type, event.Data)
		case "server-stop":
			fmt.Println("server closed!")
			cancel()
		default: // no event name
			fmt.Printf("%s: %s\n", event.Type, event.Data)
		}
	})

	if err := conn.Connect(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
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
