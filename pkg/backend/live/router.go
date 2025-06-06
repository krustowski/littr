// The package to provide a server-side instance for the SSE message streaming.
package live

import (
	chi "github.com/go-chi/chi/v5"
)

func NewLiveRouter() chi.Router {
	r := chi.NewRouter()

	// Mount the Streamer to /live API route. Wrap the SSE handler in the CORS wrapper.
	//r.Mount("/", cors(Streamer))
	r.Mount("/", Streamer)

	// Run the keepalive pacemaker.
	go beat()

	return r
}
