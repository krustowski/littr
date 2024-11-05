// The very app statistics routes and controllers logic package for the backend.
package stats

import (
	chi "github.com/go-chi/chi/v5"
)

func NewStatRouter(statController *StatController) chi.Router {
	r := chi.NewRouter()

	r.Get("/", statController.GetAll)

	return r
}
