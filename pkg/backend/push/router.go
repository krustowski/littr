// (Web)Push notifications and subscriptions routes and controllers logic package for the backend.
package push

import (
	chi "github.com/go-chi/chi/v5"
)

func NewPushRouter(pushController *PushController) chi.Router {
	r := chi.NewRouter()

	// Notification sender.
	r.Post("/", pushController.SendNotification)

	// Subscription-related routes.
	r.Post("/subscriptions", pushController.Create)
	r.Patch("/subscriptions/{uuid}", pushController.Update)
	r.Delete("/subscriptions/{uuid}", pushController.Delete)

	return r
}
