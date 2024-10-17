package auth

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", authHandler)
	r.Post("/logout", logoutHandler)

	return r
}
