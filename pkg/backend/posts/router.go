package posts

import (
	chi "github.com/go-chi/chi/v5"
	sse "github.com/alexandrevicenzi/go-sse"
)

var Streamer *sse.Server

func Router() chi.Router {
	r := chi.NewRouter()

	streamer = sse.NewServer(&sse.Options{
		Logger: nil,
	})

	// Get live flow event stream
	//
	//  @Summary      Get live flow event stream
	//  @Description  get live flow event stream
	//  @Tags         flow
	//  @Accept       json
	//  @Produce      json
	//  @Success      200  {object} octet-stream
	//  @Failure      500  {object} octet-stream
	//  @Router       /flow/live [get]
	go func() {
		for {
			streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("heartbeat"))
			time.Sleep(time.Second * 20)
		}
	}()

	r.Route("/", func(r chi.Router) {
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

	return r
}
