package stats

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getStats)

	return r
}
