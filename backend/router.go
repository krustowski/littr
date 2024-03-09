package backend

import (
	"net/http"
	"time"

	sse "github.com/alexandrevicenzi/go-sse"
	chi "github.com/go-chi/chi/v5"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	resp := response{
		Message: "litter-go API service",
		Code:    http.StatusOK,
	}
	resp.Write(w)
}

var streamer *sse.Server

// the very main API router
func LoadAPIRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(authMiddleware)

	streamer = sse.NewServer(&sse.Options{
		Logger: nil,
	})
	//defer streamer.Shutdown()

	// unauth zone (skipped at auth)
	r.Get("/", rootHandler)
	//r.Post("/auth", authHandler)
	r.Route("/auth", func(r chi.Router) {
		r.Post("/", authHandler)
		r.Post("/password", resetHandler)
	})
	r.Get("/dump", dumpHandler)

	go func() {
		for {
			streamer.SendMessage("/api/flow/live", sse.SimpleMessage("heartbeat"))
			time.Sleep(time.Second * 30)
		}
	}()

	r.Route("/flow", func(r chi.Router) {
		r.Get("/", getPosts)
		// ->backend/streamer.go
		r.Mount("/live", streamer)
		// single-post view request
		r.Route("/post", func(r chi.Router) {
			r.Get("/{postNo}", getSinglePost)
		})
		// user flow page request
		r.Route("/user", func(r chi.Router) {
			r.Get("/{nick}", getUserPosts)
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
		r.Delete("/", deleteSubscription)
		//r.Get("/vapid", generateVapidKeyPair)
	})

	r.Get("/stats", statsHandler)

	r.Route("/users", func(r chi.Router) {
		r.Get("/", getUsers)
		r.Post("/", addNewUser)
		r.Put("/", updateUser)
		r.Delete("/", deleteUser)
	})

	return r
}
