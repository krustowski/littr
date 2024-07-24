package users

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", getUsers)
		r.Post("/", addNewUser)
		r.Patch("/passphrase", resetHandler)

		r.Get("/{nickname}", getOneUser)
		//r.Put("/{nickname}", updateUser)
		r.Delete("/{nickname}", deleteUser)

		r.Post("/{nickname}/avatar", postUsersAvatar)
		r.Get("/{nickname}/posts", getUserPosts)
		r.Post("/{nickname}/request", addToRequestList)
		r.Delete("/{nickname}/request", removeFromRequestList)
		r.Patch("/{nickname}/lists", updateUserList)
		r.Patch("/{nickname}/options", updateUserOption)
		r.Patch("/{nickname}/passphrase", updateUserPassphrase)
		//r.Patch("/{nickname}/private", togglePrivateMode)
		//r.Patch("/{nickname}/localtime", toggleLocalTimeMode)
	})

	return r
}
