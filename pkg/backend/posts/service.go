package posts

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/models"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
)

//
//  models.PostServiceInterface implementation
//

type postService struct {
	notifService   models.NotificationServiceInterface
	postRepository models.PostRepositoryInterface
	userRepository models.UserRepositoryInterface
}

func NewPostService(
	notifService models.NotificationServiceInterface,
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.PostServiceInterface {
	if notifService == nil || postRepository == nil || userRepository == nil {
		return nil
	}

	return &postService{
		notifService:   notifService,
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *postService) Create(ctx context.Context, post *models.Post) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	// To patch loaded invalid user data from LocalStorage = the caller's Nickname now.
	if post.Nickname == "" {
		post.Nickname = callerID
	}

	// The caller must be the author of such post to be added.
	if post.Nickname != callerID {
		return fmt.Errorf(common.ERR_POSTER_INVALID)
	}

	// Deny blank post.
	if post.Content == "" && post.Figure == "" && post.Data == nil {
		return fmt.Errorf(common.ERR_POST_BLANK)
	}

	//
	//  Preprocess the post
	//

	// Prepare the timestamps to sigh the post.
	timestampFull := time.Now()
	timestampUnix := strconv.FormatInt(timestampFull.UnixNano(), 10)

	post.ID = timestampUnix
	post.Timestamp = timestampFull

	// Compose a payload for the image processing.
	imagePayload := &image.ImageProcessPayload{
		ImageByteData: &post.Data,
		ImageFileName: post.Figure,
		ImageBaseName: post.ID,
	}

	// Uploaded figure handling.
	if post.Data != nil && post.Figure != "" {
		imgReference, err := image.ProcessImageBytes(imagePayload)
		if err != nil {
			return err
		}

		// Ensure the image reference pointer is a valid one.
		if imgReference != nil {
			post.Figure = *imgReference
		}

		post.Data = make([]byte, 0)
	}

	//
	//  Validation end --- dispatch the post to repository.
	//

	if err := s.postRepository.Save(post); err != nil {
		return fmt.Errorf("%s: %s", common.ERR_POST_SAVE_FAIL, err.Error())
	}

	// Find matches using regexp compiling to '@username' matrix
	re := regexp.MustCompile(`@(?P<text>\w+)`)
	matches := re.FindAllStringSubmatch(post.Content, -1)

	for _, match := range matches {
		receiverName := match[1]

		// Fetch related data from the database
		receiver, err := s.userRepository.GetByID(receiverName)
		if err != nil {
			continue
		}

		// Do not notify the same person --- OK condition
		if receiverName == callerID {
			continue
		}

		// Do not notify user --- notifications disabled --- OK condition
		if len(receiver.Devices) == 0 {
			continue
		}

		// Compose the body of this notification
		body, err := json.Marshal(app.Notification{
			Title: "littr mention",
			Icon:  "/web/apple-touch-icon.png",
			Body:  callerID + " mentioned you in their post",
			Path:  "/flow/posts/" + post.ID,
		})
		if err != nil {
			//fmt.Printf(common.ERR_PUSH_BODY_COMPOSE_FAIL)
			continue
		}

		opts := &push.NotificationOpts{
			Receiver: receiverName,
			Devices:  &receiver.Devices,
			Body:     &body,
			Repo:     s.userRepository,
		}

		// Send the webpush notification(s)
		push.SendNotificationToDevices(opts)
	}

	if post.ReplyToID != "" {
		if err := s.notifService.SendNotification(ctx, post.ReplyToID); err != nil {
			fmt.Print(err.Error())
		}
	}

	// Broadcast the new post event.
	live.BroadcastMessage(live.EventPayload{Data: "post," + post.Nickname, Type: "message"})

	return nil
}

func (s *postService) Update(ctx context.Context, post *models.Post) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

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

func (s *postService) Delete(ctx context.Context, postID string) error {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

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

func (s *postService) FindAll(ctx context.Context, pageOpts interface{}) (*map[string]models.Post, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

	req, ok := pageOpts.(*PostPagingRequest)
	if !ok {
		return nil, nil, fmt.Errorf(common.ERR_REQUEST_TYPE_UNKNOWN)
	}

	// Compose a pagination options object to paginate posts.
	opts := &pages.PageOptions{
		CallerID: callerID,
		PageNo:   req.PageNo,
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

func (s *postService) FindPage(ctx context.Context, opts interface{}) (*map[string]models.Post, *map[string]models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

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

func (s *postService) FindByID(ctx context.Context, postID string) (*models.Post, *models.User, error) {
	// Fetch the caller's ID from the context.
	callerID := common.GetCallerID(ctx)

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
