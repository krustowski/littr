package posts

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.PostRepositoryInterface.
type PostRepository struct {
	cache db.Cacher
}

func NewPostRepository(cache db.Cacher) *PostRepository {
	if cache == nil {
		return nil
	}

	return &PostRepository{
		cache: cache,
	}
}

func (r *PostRepository) GetAll() (*map[string]models.Post, error) {
	rawPosts, count := r.cache.Range()
	if count == 0 {
		return nil, fmt.Errorf("no items found")
	}

	posts := make(map[string]models.Post)

	// Assert types to fetched interface map.
	for key, rawPost := range *rawPosts {
		post, ok := rawPost.(models.Post)
		if !ok {
			return nil, fmt.Errorf("post's data corrupted")
		}

		posts[key] = post
	}

	return &posts, nil
}

func (r *PostRepository) GetByID(postID string) (*models.Post, error) {
	// Fetch the post from the cache.
	rawPost, found := r.cache.Load(postID)
	if !found {
		return nil, fmt.Errorf("requested post not found")
	}

	// Assert the type
	post, ok := rawPost.(models.Post)
	if !ok {
		return nil, fmt.Errorf("post's data corrupted")
	}

	return &post, nil
}

func (r *PostRepository) Save(post *models.Post) error {
	// Store the post using its key in the cache.
	saved := r.cache.Store(post.ID, *post)
	if !saved {
		return fmt.Errorf("an error occurred while saving a post")
	}

	return nil
}

func (r *PostRepository) Delete(postID string) error {
	// Simple post's deleting.
	deleted := r.cache.Delete(postID)
	if !deleted {
		return fmt.Errorf("post data could not be purged from the database")
	}

	return nil
}
