package posts

import (
	"context"
	"fmt"

	//"strconv"
	//"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"

	//"go.vxn.dev/littr/pkg/helpers"
	"go.vxn.dev/littr/pkg/models"
)

//
// models.PostServiceInterface implementation
//

type PostService struct {
	postRepository models.PostRepositoryInterface
	userRepository models.UserRepositoryInterface
}

func NewPostService(
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.PostServiceInterface {
	if postRepository == nil || userRepository == nil {
		return nil
	}

	return &PostService{
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *PostService) Create(ctx context.Context, post *models.Post) error {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// To patch loaded invalid user data from LocalStorage = the caller's Nickname now.
	if post.Nickname == "" {
		post.Nickname = callerID
	}

	// The caller must be the author of such post to be added.
	if post.Nickname != callerID {
		return fmt.Errorf(common.ERR_POSTER_INVALID)
	}

	//
	//  Validation end --- dispatch the post to repository.
	//

	if err := s.postRepository.Save(post); err != nil {
		return fmt.Errorf("%s: %s", common.ERR_POST_SAVE_FAIL, err.Error())
	}

	// Broadcast the new post event.
	live.BroadcastMessage(live.EventPayload{Data: "post," + post.ID, Type: "message"})

	return nil
}

func (s *PostService) Update(ctx context.Context, post *models.Post) error {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the actual post to verify its content to be updated..
	dbPost, err := s.postRepository.GetByID(post.ID)
	if err != nil {
		return err
	}

	// Check the post's ownership. The author cannot vote on such post.
	if dbPost.Nickname == callerID {
		return fmt.Errorf(common.ERR_POSTER_INVALID)
	}

	// Save the changes in repository.
	return s.postRepository.Save(dbPost)
}

func (s *PostService) Delete(ctx context.Context, postID string) error {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the actual post to verify it can be deleted at all.
	post, err := s.postRepository.GetByID(postID)
	if err != nil {
		return err
	}

	// Check the post's ownership.
	if post.Nickname != callerID {
		return fmt.Errorf(common.ERR_POLL_DELETE_FOREIGN)
	}

	// Try to delete the post.
	return s.postRepository.Delete(postID)
}

func (s *PostService) FindAll(ctx context.Context) (*map[string]models.Post, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the pageNo from the context.
	pageNo, ok := ctx.Value("pageNo").(int)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_PAGENO_INCORRECT)
	}

	// Compose a pagination options object to paginate posts.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			Plain: true,
		},
	}

	// Request the page of posts from the post repository.
	posts, _, err := s.postRepository.GetPage(opts)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	// Patch the user's data for export.
	patchedCaller := (*common.FlushUserData(&map[string]models.User{callerID: *caller}, callerID))[callerID]

	return posts, &patchedCaller, nil
}

func (s *PostService) FindPage(ctx context.Context, opts interface{}) (*map[string]models.Post, *map[string]models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Request the page of posts from the post repository.
	posts, users, err := s.postRepository.GetPage(opts)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	// Assign the caller into the users map.
	(*users)[callerID] = *caller

	// Patch the user's data for export.
	return posts, common.FlushUserData(users, callerID), nil
}

func (s *PostService) FindByID(ctx context.Context, postID string) (*models.Post, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_CALLER_FAIL)
	}

	// Fetch the post.
	post, err := s.postRepository.GetByID(postID)
	if err != nil {
		return nil, nil, err
	}

	// Request the caller from the user repository.
	caller, err := s.userRepository.GetByID(callerID)
	if err != nil {
		return nil, nil, err
	}

	// Patch the user's data for export.
	patchedCaller := (*common.FlushUserData(&map[string]models.User{callerID: *caller}, callerID))[callerID]

	return post, &patchedCaller, nil
}
