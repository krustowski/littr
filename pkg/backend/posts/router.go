// Posts routes and controllers logic package for the backend.
package posts

import (
	chi "github.com/go-chi/chi/v5"
)

func NewPostRouter(postController *PostController) chi.Router {
	r := chi.NewRouter()

	r.Get("/", postController.GetAll)
	r.Post("/", postController.Create)

	// single-post view request
	r.Get("/{postID}", postController.GetByID)

	// user flow page request -> backend/users/controllers.go
	/*r.Route("/user", func(r chi.Router) {
		r.Get("/{nick}", getUserPosts)
	})*/

	r.Patch("/{postID}/star", postController.UpdateReactions)
	r.Delete("/{postID}", postController.Delete)

	r.Get("/hashtags/{hashtag}", postController.GetByHashtag)

	return r
}
