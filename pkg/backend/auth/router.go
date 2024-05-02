package auth

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Post("/", authHandler)
		r.Post("/logout", logoutHandler)
	})

	return r
}
