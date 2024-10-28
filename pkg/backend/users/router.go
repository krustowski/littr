// Users routes and controllers logic package for the backend.
package users

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	// Basic routes handlers.
	r.Get("/", getUsers)
	r.Post("/", addNewUser)

	// Handler for the user activation.
	r.Post("/activation/{uuid}", activationRequestHandler)

	// Passphrase-related routes handlers.
	r.Post("/passphrase/request", resetRequestHandler)
	r.Post("/passphrase/reset", resetPassphraseHandler)

	// User getter handlers.
	r.Get("/{userID}", getOneUser)
	r.Get("/caller", getOneUser)

	// User modification/deletion handlers.
	//r.Put("/{nickname}", updateUser)
	r.Delete("/{userID}", deleteUser)

	// User's settings modification routes handlers.
	r.Post("/{userID}/avatar", postUserAvatar)
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
