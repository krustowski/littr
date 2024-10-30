// Runtime profiling metapackage.
package pprof

import (
	prf "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//  Sources:
//
//  https://pkg.go.dev/net/http/pprof
//  https://github.com/go-chi/chi/blob/master/middleware/profiler.go

// The pprof profiler common router.
func Router() chi.Router {
	r := chi.NewRouter()

	// Do not cache the profiles.
	r.Use(middleware.NoCache)

	r.Get("/pprof", prf.Index)
	r.Get("/pprof/cmdline", prf.Cmdline)
	r.Get("/pprof/profile", prf.Profile)
	r.Get("/pprof/trace", prf.Trace)

	r.Mount("/pprof/allocs", prf.Handler("allocs"))
	r.Mount("/pprof/block", prf.Handler("block"))
	r.Mount("/pprof/goroutine", prf.Handler("goroutine"))
	r.Mount("/pprof/heap", prf.Handler("heap"))
	r.Mount("/pprof/mutex", prf.Handler("mutex"))
	r.Mount("/pprof/threadcreate", prf.Handler("threadcreate"))

	return r
}
