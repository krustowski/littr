// Users routes and controllers logic package for the backend.
package users

import (
	chi "github.com/go-chi/chi/v5"
)

func NewUserRouter(userController *UserController) chi.Router {
	r := chi.NewRouter()

	// Basic routes handlers.
	r.Get("/", userController.GetAll)
	r.Post("/", userController.Create)

	// Handler for the user activation.
	r.Post("/activation", userController.Activate)

	// Passphrase-related routes handlers.
	r.Post("/passphrase/request", userController.PassphraseResetRequest)
	r.Post("/passphrase/reset", userController.PassphraseReset)

	// User getter handlers.
	r.Get("/{userID}", userController.GetByID)
	//r.Get("/caller", userController.GetByID)

	// User modification/deletion handlers.
	r.Delete("/{userID}", userController.Delete)

	// User's settings modification routes handlers.
	r.Post("/{userID}/avatar", userController.UploadAvatar)
	r.Get("/{userID}/posts", userController.GetPosts)

	r.Patch("/{userID}/lists", userController.UpdateLists)
	r.Patch("/{userID}/options", userController.UpdateOptions)
	r.Patch("/{userID}/passphrase", userController.UpdatePassphrase)

	r.Post("/{userID}/subscriptions", userController.Subscribe)
	r.Patch("/{userID}/subscriptions/{uuid}", userController.UpdateSubscription)
	r.Delete("/{userID}/subscriptions/{uuid}", userController.Unsubscribe)

	return r
}
