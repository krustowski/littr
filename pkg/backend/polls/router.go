// Polls routes and controllers logic package for the backend.
package polls

import (
	chi "github.com/go-chi/chi/v5"
)

func NewPollRouter(pollController *PollController) chi.Router {
	r := chi.NewRouter()

	// Route routes.
	r.Get("/", pollController.GetAll)
	r.Post("/", pollController.Create)

	// Operations on an existing resource.
	r.Get("/{pollID}", pollController.GetByID)
	r.Patch("/{pollID}", pollController.Update)
	r.Patch("/{pollID}", pollController.Update)
	r.Delete("/{pollID}", pollController.Delete)

	return r
}
