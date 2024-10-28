// Posts routes and controllers logic package for the backend.
package posts

import (
	chi "github.com/go-chi/chi/v5"
)

func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", getPosts)
	r.Post("/", addNewPost)

	// single-post view request
	r.Get("/{postID}", getSinglePost)

	// user flow page request -> backend/users/controllers.go
	/*r.Route("/user", func(r chi.Router) {
		r.Get("/{nick}", getUserPosts)
	})*/

	r.Patch("/{postID}/star", updatePostStarCount)
	r.Delete("/{postID}", deletePost)

	r.Get("/hashtag/{hashtag}", fetchHashtaggedPosts)

	return r
}
