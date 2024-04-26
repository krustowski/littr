package push

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/vapid", fetchVAPIDKey)
		r.Post("/subscription", subscribeToNotifications)
		r.Post("/notification", sendNotification)
		r.Delete("/{nickname}", deleteSubscription)
	})

	return r
}
