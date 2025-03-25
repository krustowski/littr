//go:build !wasm || server
// +build !wasm server

package main

import (
	//"compress/flate"
	"context"
	"errors"
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

type server struct {
	ctx context.Context

	db db.DatabaseKeeper

	l common.Logger

	listener net.Listener

	once sync.Once

	// The very main HTTP server's struct pointer.
	srv *http.Server

	// The WaitGroup for the graceful HTTP server shutdown.
	wg *sync.WaitGroup
}

func newServer() App {
	return &server{}
}

func (s *server) Run() {
	s.init()
	s.handleSignalsShutdown()

	s.setupRouterServer()
	s.serve()

	//
	//  Shutdown
	//

	// Wait for the graceful HTTP server shutdown attempt.
	s.wg.Wait()

	// This is the final log before the application exits for real! Reset the timer not to log the whole server's uptime.
	// https://dev.to/mokiat/proper-http-shutdown-in-go-3fji
	s.l.ResetTimer().Msg("the HTTP server has stopped serving new connections, exit").Log()

	//defer os.Exit(0)
	//return
}

func (s *server) init() {
	s.once.Do(func() {
		s.l = common.NewLogger(nil, "initServer")
		s.l.Msg("server preflight checks start").Log()

		if config.ServerSecret == "" || config.DataDumpToken == "" {
			panic(errMissingSecretOrToken)
		}

		s.db = db.NewDatabase()

		// Lock the database stack for read, unlock it for write (see pkg/backend/db/init.go for more).
		s.db.LockRead()

		// Load the persistent data from the filesystem to memory.
		s.l.Msg("dumped load result: " + db.LoadAll()).Status(http.StatusOK).Log()

		// Unlock the read access.
		s.db.UnlockRead()

		//
		//  Database and data initialization (caches themselves and the database state is initialized on pkg db import).
		//

		// Run data migration procedures to the database schema.
		migrationsReport := db.RunMigrations()

		migrationsStatus := func() int {
			if strings.Contains(migrationsReport, "false") {
				return http.StatusInternalServerError
			}
			return http.StatusOK
		}()

		l.Msg(migrationsReport).Status(migrationsStatus).Log()

	})

}

func (s *server) handleSignalsShutdown() {
	// Handle system calls and signals to properly shutdown the server.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// The signals monitoring goroutine.
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Wait for signals.
		sig := <-sigs
		signal.Stop(sigs)

		// Create a shutdown logger.
		l := common.NewLogger(nil, "shutdown")

		// Log and broadcast the message that the server is to shutdown.
		l.Msg("trap signal: " + sig.String() + ", stopping the HTTP server gracefully...").Log()

		live.BroadcastMessage(live.EventPayload{Data: "server-stop", Type: "message"})
		live.BroadcastMessage(live.EventPayload{Data: "server-stop", Type: "close"})

		// "Lock" the write access to the database. <--- causes threadlock and app exit deferals when used with the actual lock !!!
		s.db.LockWrite()

		// Dump all in-memory databases.
		report, err := s.db.DumpAll()
		if err != nil {
			l.Error(err).Log()
		} else {
			l.Msg(report).Log()
		}

		// Release the lock, but keep the database read-only. The lock blocks the main thread and defers the application shutdown.
		s.db.ReleaseLock()

		// Fetch a context to send to gracefully shutdown the HTTP server.
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Terminate the SSE server. Method Shutdown below implicitly shuts down the SSE Provider.
		if live.Streamer != nil {
			if err := live.Streamer.Shutdown(sctx); err != nil {
				l.Error(err).Log()
			}
		}

		// Terminate the HTTP server from here, give it 5 seconds to shutdown gracefully..
		if err := s.srv.Shutdown(sctx); err != nil {
			l.Error(errServerShutdownFailed, err).Log()

			// Force terminate the HTTP server if failed to stop gracefully.
			if err := s.srv.Close(); err != nil {
				l.Error(err).Log()
			}
			return
		}

		l.Msg("graceful shutdown complete").Log()
		// The graceful end of the goroutine = the program is about to exit.
	}()
}

func (s *server) setupRouterServer() {
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
	var err error
	if s.listener, err = net.Listen("tcp", ":"+config.ServerPort); err != nil {
		// Cannot listen on such address = a permission issue?
		panic(err)
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

	// Register the (mostly) cache metrics.
	metrics.RegisterAll()

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

	// Create a custom HTTP server. WriteTimeout is set to 0 (infinite) due to the SSE subserver present.
	s.srv = &http.Server{
		Addr: s.listener.Addr().String(),
		//ReadTimeout: 0 * time.Second,
		WriteTimeout: 0 * time.Second,
		Handler:      r,
	}
}

func (s *server) serve() {
	//
	//  Start the server
	//

	s.l.Msg("init done, starting the HTTP server (v" + config.AppVersion).Log()

	defer func() {
		if err := s.listener.Close(); err != nil {
			s.l.Error(err).Log()
		}
	}()

	// Send the SSE regarding the server start.
	go func() {
		time.Sleep(time.Second * 30)
		live.BroadcastMessage(live.EventPayload{Data: "server-start", Type: "message"})
	}()

	// Inject the logger to the connection context.
	s.srv.ConnContext = func(ctx context.Context, c net.Conn) context.Context {
		if ctx == nil {
			ctx = context.Background()
		}

		if c == nil {
			return nil
		}

		return context.WithValue(ctx, common.LoggerContextKey, s.l)
	}

	// Start serving using the created net listener.
	if err := s.srv.Serve(s.listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// Reset the timer not to log the whole server's uptime.
		s.l.ResetTimer().Error(err).Log()
	}
}
