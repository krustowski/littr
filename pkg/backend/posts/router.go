package posts

import (
	"time"

	sse "github.com/alexandrevicenzi/go-sse"
	chi "github.com/go-chi/chi/v5"
)

var Streamer *sse.Server

func Router() chi.Router {
	r := chi.NewRouter()

	Streamer = sse.NewServer(&sse.Options{
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
			Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("heartbeat"))
			time.Sleep(time.Second * 20)
		}
	}()

	r.Route("/", func(r chi.Router) {
		r.Get("/", getPosts)
		r.Post("/", addNewPost)
		// ->backend/streamer.go
		r.Mount("/live", Streamer)
		// single-post view request
		r.Get("/{postID}", getSinglePost)
		// user flow page request
		/*r.Route("/user", func(r chi.Router) {
			r.Get("/{nick}", getUserPosts)
		})*/
		//r.Put("/", updatePost)
		r.Patch("/{postID}/star", updatePostStarCount)
		r.Delete("/{postID}", deletePost)
	})

	return r
}
