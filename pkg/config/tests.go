package config

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Create a custom network TCP connection listener.
func PrepareTestListener(t *testing.T) net.Listener {
	if t == nil {
		return nil
	}

	return PrepareTestListenerWithPort(t, DEFAULT_TEST_PORT)
}

func PrepareTestListenerWithPort(t *testing.T, port string) net.Listener {
	if t == nil {
		return nil
	}

	if port == "" {
		t.Errorf("listener's port not specified")
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		// Cannot listen on such address = a permission issue or already used
		t.Errorf(err.Error())
		return nil
	}

	return listener
}

// Create a custom HTTP server configuration suitable to serve with the SSE streamer.
func PrepareTestServer(t *testing.T, listener net.Listener, handler http.Handler) *httptest.Server {
	if t == nil || listener == nil || handler == nil {
		return nil
	}

	// Common HTTP server config.
	serverConfig := &http.Server{
		Addr: listener.Addr().String(),
		//ReadTimeout: 0 * time.Second,
		WriteTimeout: 0 * time.Second,
		Handler:      handler,
	}

	// Test HTTP server config.
	testServer := &httptest.Server{
		Listener:    listener,
		EnableHTTP2: false,
		Config:      serverConfig,
	}

	return testServer
}
