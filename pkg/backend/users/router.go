package users

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	// basic routes
	r.Get("/", getUsers)
	r.Post("/", addNewUser)

	// passphrase-related routes
	r.Post("/passphrase/request", resetRequestHandler)
	r.Post("/passphrase/reset", resetPassphraseHandler)

	// incomplete CRUD
	r.Get("/{userID}", getOneUser)
	r.Get("/caller", getOneUser)
	//r.Put("/{nickname}", updateUser)
	r.Delete("/{userID}", deleteUser)

	// user's settings routes
	r.Post("/{userID}/avatar", postUsersAvatar)
	r.Get("/{userID}/posts", getUserPosts)
	r.Patch("/{userID}/lists", updateUserList)
	r.Patch("/{userID}/options", updateUserOption)
	r.Patch("/{userID}/passphrase", updateUserPassphrase)

	// request-to-follow routes
	// (depraceted, use /{userID}/lists route instead)
	//r.Post("/{userID}/request", addToRequestList)
	//r.Delete("/{userID}/request", removeFromRequestList)

	return r
}
