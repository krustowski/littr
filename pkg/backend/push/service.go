package push

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/maxence-charriere/go-app/v10/pkg/app"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/models"
)

type SubscriptionService struct {
	postRepository         models.PostRepositoryInterface
	subscriptionRepository models.SubscriptionRepositoryInterface
}

func NewSubscriptionService(
	postRepository models.PostRepositoryInterface,
	subscriptionRepository models.SubscriptionRepositoryInterface,
) *SubscriptionService {

	if postRepository == nil || subscriptionRepository == nil {
		return nil
	}

	return &SubscriptionService{
		postRepository:         postRepository,
		subscriptionRepository: subscriptionRepository,
	}
}

func (s *SubscriptionService) SendNotification(ctx context.Context, postID string) error {
	// Fetch the callerID from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the caller's ID")
	}

	if postID == "" {
		return fmt.Errorf(common.ERR_POSTID_BLANK)
	}

	post, err := s.postRepository.GetByID(postID)
	if err != nil {
		return err
	}

	dbSub, err := s.subscriptionRepository.GetByUserID(post.Nickname)
	if err != nil {
		// It is OK for this to return nothing, loop can handle it later...
		//return err
	}

	// Do not notify the same person --- OK condition.
	if post.Nickname == callerID {
		return nil
	}

	// Do not notify such user --- notifications disabled --- OK condition.
	if dbSub == nil || len(*dbSub) == 0 {
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
		Devices:  dbSub,
		Body:     &body,
		//Logger:   l,
	}

	// Send the webpush notification(s).
	SendNotificationToDevices(opts)

	return nil
}
