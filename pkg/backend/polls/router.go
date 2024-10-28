// Polls routes and controllers logic package for the backend.
package polls

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", getPolls)
		r.Post("/", addNewPoll)
		r.Get("/{pollID}", getSinglePoll)
		r.Put("/{pollID}", updatePoll)
		r.Delete("/{pollID}", deletePoll)
	})

	return r
}
