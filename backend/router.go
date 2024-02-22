package backend

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{
		Message: "litter-go API service",
		Code:    http.StatusOK,
	}
	resp.Write(w)
}

// the very main API router
func LoadAPIRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(authMiddleware)

	// unauth zone (skipped at auth)
	r.Get("/", rootHandler)
	r.Post("/auth", authHandler)
	r.Get("/dump", dumpHandler)

	r.Get("/stats", statsHandler)

	r.Route("/flow", func(r chi.Router) {
		r.Get("/", getPosts)
		//r.Get("/{pageNo}", getPosts)
		// user flow page request
		r.Route("/user", func(r chi.Router) {
			r.Get("/{nick}", getUserPosts)
		})
		// single-post view request
		r.Route("/post", func(r chi.Router) {
			r.Get("/{postNo}", getSinglePost)
		})
		r.Post("/", addNewPost)
		//r.Put("/", updatePost)
		r.Put("/star", updatePostStarCount)
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
