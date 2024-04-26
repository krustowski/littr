package users

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", getUsers)
		r.Post("/", addNewUser)
		r.Post("/{nickname}/reset", resetPassphrase)
		r.Get("/{nickname}", getOneUser)
		r.Put("/{nickname}", updateUser)
		r.Delete("/{nickname}", deleteUser)
	})

	return r
}
