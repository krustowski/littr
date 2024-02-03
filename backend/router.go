package backend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("litter-go API root"))
}

// the very main API router
func LoadAPIRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rootHandler)
	r.Post("/auth", authHandler)
	r.Get("/dump", dumpHandler)
	r.Get("/stats", statsHandler)

	r.Route("/flow", func(r chi.Router) {
		r.Get("/", getPosts)
		r.Get("/{page}", getPostsPaged)
		r.Post("/", addNewPost)
		r.Put("/", updatePost)
		r.Delete("/", deletePost)
	})
	r.Route("/polls", func(r chi.Router) {
		r.Get("/", getPolls)
		r.Post("/", addNewPoll)
		r.Put("/", updatePoll)
		r.Delete("/", deletePoll)
	})
	r.Route("/push", func(r chi.Router) {
		r.Post("/", subscribeToNotifs)
		r.Put("/", sendNotif)
	})
	r.Route("/users", func(r chi.Router) {
		r.Get("/", getUsers)
		r.Post("/", addNewUser)
		r.Put("/", updateUser)
		r.Delete("/", deleteUser)
	})

	return r
}
