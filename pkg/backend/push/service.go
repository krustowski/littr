package push

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type notificationService struct {
	postRepository models.PostRepositoryInterface
	userRepository models.UserRepositoryInterface
}

func NewNotificationService(
	postRepository models.PostRepositoryInterface,
	userRepository models.UserRepositoryInterface,
) models.NotificationServiceInterface {

	if postRepository == nil || userRepository == nil {
		return nil
	}

	return &notificationService{
		postRepository: postRepository,
		userRepository: userRepository,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, postID string) error {
	// Fetch the callerID from the given context.
	callerID := common.GetCallerID(ctx)

	if postID == "" {
		return fmt.Errorf(common.ERR_POSTID_BLANK)
	}

	post, err := s.postRepository.GetByID(postID)
	if err != nil {
		return err
	}

	user, err := s.userRepository.GetByID(post.Nickname)
	if err != nil {
		return err
	}

	// Do not notify the same person --- OK condition.
	if post.Nickname == callerID {
		return nil
	}

	// Do not notify such user --- notifications disabled --- OK condition.
	if len(user.Devices) == 0 {
		return nil
	}

	// Compose the body of this notification.
	body, _ := json.Marshal(app.Notification{
		Title: "littr reply",
		Icon:  "/web/apple-touch-icon.png",
		Body:  callerID + " replied to your post",
		Path:  "/flow/posts/" + post.ID,
	})

	opts := &NotificationOpts{
		Receiver: post.Nickname,
		Devices:  &user.Devices,
		Body:     &body,
	}

	// Send the webpush notification(s).
	SendNotificationToDevices(opts)

	return nil
}
