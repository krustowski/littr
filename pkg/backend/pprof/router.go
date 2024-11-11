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
func NewRouter() chi.Router {
	r := chi.NewRouter()

	// Do not cache the profiles.
	r.Use(middleware.NoCache)

	r.Get("/", prf.Index)
	r.Get("/cmdline", prf.Cmdline)
	r.Get("/profile", prf.Profile)
	r.Get("/trace", prf.Trace)

	r.Mount("/allocs", prf.Handler("allocs"))
	r.Mount("/block", prf.Handler("block"))
	r.Mount("/goroutine", prf.Handler("goroutine"))
	r.Mount("/heap", prf.Handler("heap"))
	r.Mount("/mutex", prf.Handler("mutex"))
	r.Mount("/threadcreate", prf.Handler("threadcreate"))

	return r
}
