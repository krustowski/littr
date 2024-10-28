// (Web)Push notifications and subscriptions routes and controllers logic package for the backend.
package push

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	// public VAPID key fetcher
	r.Get("/vapid", fetchVAPIDKey)

	// subscription-related routes
	r.Post("/subscription", subscribeToNotifications)
	r.Put("/subscription/{uuid}/mention", updateSubscription)
	r.Put("/subscription/{uuid}/reply", updateSubscription)
	r.Delete("/subscription/{uuid}", deleteSubscription)

	// notification sender
	r.Post("/notification/{postID}", sendNotification)

	return r
}
