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
	r.Post("/activation/{uuid}", userController.Activate)

	// Passphrase-related routes handlers.
	r.Post("/passphrase/{requestType}", userController.PassphraseReset)
	//r.Post("/passphrase/request", userController.ResetPassphrase)
	//r.Post("/passphrase/reset", userController.ResetPassphrase)

	// User getter handlers.
	r.Get("/{userID}", userController.GetByID)
	//r.Get("/caller", userController.GetByID)

	// User modification/deletion handlers.
	r.Delete("/{userID}", userController.Delete)

	// User's settings modification routes handlers.
	r.Post("/{userID}/avatar", userController.UploadAvatar)
	r.Get("/{userID}/posts", userController.GetPosts)

	r.Patch("/{userID}/{updateType}", userController.Update)
	//r.Patch("/{userID}/lists", userController.UpdateLists)
	//r.Patch("/{userID}/options", userController.UpdateOptions)
	//r.Patch("/{userID}/passphrase", userController.UpdatePassphrase)

	return r
}
