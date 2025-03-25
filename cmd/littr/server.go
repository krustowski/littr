//go:build !wasm || server
// +build !wasm server

package main

import (
	//"compress/flate"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	be "go.vxn.dev/littr/pkg/backend"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/metrics"
	"go.vxn.dev/littr/pkg/backend/pprof"
	"go.vxn.dev/littr/pkg/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler and its ServerHTTP method is a simple implementation of the http.Handler interface. It can be used to wrap various HTTP handlers.
// https://github.com/go-chi/chi/blob/master/_examples/custom-handler/main.go
type Handler func(w http.ResponseWriter, r *http.Request) error

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		// Handle returned error here: write it out to client.
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}

var (
	// The very main HTTP server's struct pointer.
	server *http.Server

	// The WaitGroup for the graceful HTTP server shutdown.
	wg sync.WaitGroup
)

// initClient initializes the router for the client web application (if run in browser for the first time).
func initClient() {
	initClientCommon()
}

func initServer() {
	// Prepare the Logger instance.
	l := common.NewLogger(nil, "initServer")
	l.Msg("littr server starting (init phase)...").Status(http.StatusOK).Log()

	//
	//  Prestart preparations
	//

	// Check if the server secrets (APP_PEPPER and API_TOKEN) are initialized and not empty.
	if os.Getenv("APP_PEPPER") == "" || os.Getenv("API_TOKEN") == "" {
		panic("Any of APP_PEPPER or API_TOKEN environment variables not specified. Could not continue the HTTP server prestart. Terminating now.")
	}

	// Handle system calls and signals to properly shutdown the server.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// The signals monitoring goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Wait for signals.
		sig := <-sigs
		signal.Stop(sigs)

		// Create a shutdown logger.
		l := common.NewLogger(nil, "shutdown")

		// Log and broadcast the message that the server is to shutdown.
		l.ResetTimer().Msg("trap signal: " + sig.String() + ", stopping the HTTP server gracefully...").Status(http.StatusOK).Log()
		live.BroadcastMessage(live.EventPayload{Data: "server-stop", Type: "message"})
		live.BroadcastMessage(live.EventPayload{Data: "server-stop", Type: "close"})

		// "Lock" the write access to the database. <--- causes threadlock and app exit deferals when used with the actual lock !!!
		db.Lock()

		// Dump all in-memory databases.
		l.ResetTimer().Msg(db.DumpAll()).Status(http.StatusOK).Log()

		// Release the lock, but keep the database read-only. The lock blocks the main thread and defers the application shutdown.
		db.ReleaseLock()

		// Fetch a context to send to gracefully shutdown the HTTP server.
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Terminate the SSE server. Method Shutdown below implicitly shuts down the SSE Provider.
		if live.Streamer != nil {
			live.Streamer.Shutdown(sctx)
		}

		// Terminate the HTTP server from here, give it 5 seconds to shutdown gracefully..
		if err := server.Shutdown(sctx); err != nil {
			l.Msg(fmt.Sprintf("HTTP server shutdown error: %s, force terminitaing the server..., ", err.Error())).Status(http.StatusInternalServerError).Log()

			// Force terminate the HTTP server if failed to stop gracefully.
			server.Close()
			return
		}

		l.Msg("graceful shutdown complete").Status(http.StatusOK).Log()
		// The graceful end of the goroutine = the program is about to exit.
	}()

	//
	//  Database and data initialization (caches themselves and the database state is initialized on pkg db import).
	//

	// Lock the database stack for read, unlock it for write (see pkg/backend/db/init.go for more).
	db.RLock()

	l.Msg("caches initialized").Status(http.StatusOK).Log()

	// Register the (mostly) cache metrics.
	metrics.RegisterAll()

	// Load the persistent data from the filesystem to memory.
	l.Msg("dumped load result: " + db.LoadAll()).Status(http.StatusOK).Log()

	// Unlock the read access.
	db.RUnlock()

	// Run data migration procedures to the database schema.
	migrationsReport := db.RunMigrations()

	migrationsStatus := func() int {
		if strings.Contains(migrationsReport, "false") {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}()

	l.Msg(migrationsReport).Status(migrationsStatus).Log()

	//
	//  Muxer, listener and server initialization
	//

	// Create a new go-chi muxer.
	r := chi.NewRouter()

	// Cleans out multiple slashes in the URI path.
	r.Use(middleware.CleanPath)

	// Ensures the muxer should survive the panic.
	r.Use(middleware.Recoverer)

	// Enable a proactive data compression.
	// https://pkg.go.dev/compress/flate
	/*compressor := middleware.NewCompressor(flate.HuffmanOnly, "application/wasm", "image/svg+xml", "image/gif")
	r.Use(compressor.Handler)*/

	// Create a custom network TCP connection listener.
	listener, err := net.Listen("tcp", ":"+config.ServerPort)
	if err != nil {
		// Cannot listen on such address = a permission issue?
		panic(err)
	}
	defer listener.Close()

	// Create a custom HTTP server. WriteTimeout is set to 0 (infinite) due to the SSE subserver present.
	server = &http.Server{
		Addr: listener.Addr().String(),
		//ReadTimeout: 0 * time.Second,
		WriteTimeout: 0 * time.Second,
		Handler:      r,
	}

	//
	//  Routes and handlers mounting
	//

	// Mount the very main API router spanning all the backend.
	r.Mount("/api/v1", be.NewAPIRouter())

	// Mount the pprof profiler router.
	r.Mount("/debug/pprof", pprof.NewRouter())

	// A workaround to serve a proper favicon icon.
	r.Method("GET", "/favicon.ico", Handler(func(w http.ResponseWriter, r *http.Request) error {
		http.ServeFile(w, r, "/opt/web/favicon.ico")
		return nil
	}))

	// Register the Prometheus metrics' handle.
	r.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		Registry: metrics.Registry,
	}))

	// Serve custom compressed client binary.
	/*r.Method("GET", "/web/app.wasm", Handler(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Type", "application/wasm")

		wasmBinary, err := os.ReadFile("/opt/web/app.wasm.gz")
		if err != nil {
			return err
		}

		w.Write(wasmBinary)
		return nil
	}))*/

	r.Handle("/*", appHandler)

	//
	//  Start the server
	//

	l.Msg("starting the HTTP server (app v" + os.Getenv("APP_VERSION") + ")...").Status(http.StatusOK).Log()

	// Send the SSE regarding the server start.
	go func() {
		time.Sleep(time.Second * 30)
		live.BroadcastMessage(live.EventPayload{Data: "server-start", Type: "message"})
	}()

	// Start serving using the created net listener.
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Reset the timer not to log the whole server's uptime.
		l.ResetTimer().Msg(fmt.Sprintf("HTTP server error: %s", err.Error())).Status(http.StatusInternalServerError).Log()
	}

	//
	//  Exit
	//

	// Wait for the graceful HTTP server shutdown attempt.
	wg.Wait()

	// This is the final log before the application exits for real! Reset the timer not to log the whole server's uptime.
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	l.ResetTimer().Msg("HTTP server has stopped serving new connections, exit").Status(http.StatusOK).Log()

	defer os.Exit(0)
	return
}
