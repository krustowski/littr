package posts

import (
	"fmt"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

// The implementation of pkg/models.PostRepositoryInterface.
type PostRepository struct {
	cache db.Cacher
}

func NewPostRepository(cache db.Cacher) models.PostRepositoryInterface {
	if cache == nil {
		return nil
	}

	return &PostRepository{
		cache: cache,
	}
}

func (r *PostRepository) GetAll(pageOpts interface{}) (*map[string]models.Post, error) {
	// Assert type for pageOptions.
	opts, ok := pageOpts.(*pages.PageOptions)
	if !ok {
		return nil, fmt.Errorf("cannot read the page options at the repository level")
	}

	// Fetch page according to the calling user (in options).
	pagePtrs := pages.GetOnePage(*opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || (*pagePtrs.Posts) == nil {
		return nil, fmt.Errorf(common.ERR_PAGE_EXPORT_NIL)
	}

	// If zero items were fetched, no need to continue asserting types.
	if len(*pagePtrs.Posts) == 0 {
		return nil, fmt.Errorf("no posts found in the database")
	}

	return pagePtrs.Posts, nil

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
