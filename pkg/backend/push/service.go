package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/helpers"
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

func (s *SubscriptionService) Create(ctx context.Context, device *models.Device) error {
	// Fetch the callerID from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the caller's ID")
	}

	// Check whether the given device is blank.
	if reflect.DeepEqual(*device, (models.Device{})) {
		return fmt.Errorf(common.ERR_DEVICE_BLANK)
	}

	// Loop over []models.Device fetched from SubscriptionRepository.
	dbSub, err := s.subscriptionRepository.GetByUserID(callerID)
	if err != nil && !errors.Is(err, ErrSubscriptionNotFound) {
		return err
	}

	// Loop through the devices and check its presence (present = already subscribed).
	if dbSub != nil {
		for _, dev := range *dbSub {
			if dev.UUID == device.UUID {
				// Found a match, thus request was fired twice, or someone tickles the API.
				return fmt.Errorf(common.ERR_DEVICE_SUBSCRIBED_ALREADY)
			}
		}

		// Append new device into the devices array for such user.
		*dbSub = append(*dbSub, *device)
	} else {
		dbSub = &[]models.Device{*device}
	}

	if err := s.subscriptionRepository.Save(callerID, dbSub); err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) Update(ctx context.Context, uuid, tagName string) error {
	// Fetch the callerID from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the caller's ID")
	}

	dbSub, err := s.subscriptionRepository.GetByUserID(callerID)
	if err != nil {
		return err
	}

	var (
		device models.Device
		found  bool
	)

	// Get the specified device by its UUID.
	for _, dev := range *dbSub {
		if dev.UUID == uuid {
			device = dev
			found = true
		}
	}

	if !found {
		return ErrSubscriptionNotFound
	}

	// Append the tag if not present already.
	if !helpers.Contains(device.Tags, tagName) {
		device.Tags = append(device.Tags, tagName)
	} else {
		// Remove the existing tag.
		tags := device.Tags

		var newTags []string

		// Loop over tags and yeet the one existing and requested.
		for _, tag := range tags {
			if tag == tagName {
				continue
			}
			newTags = append(newTags, tag)
		}

		// Update device's tags.
		device.Tags = newTags
	}

	// Loop through loaded devices and look for the requested one.
	// If found, update the existing requested device.
	found = false
	for idx, dev := range *dbSub {
		// requested dev's UUID matches loaded device's UUID
		if dev.UUID == device.UUID {
			// save the requested device to the device array
			(*dbSub)[idx] = device
			found = true
		}
	}

	// append new device into the devices array for such user, if not found before
	if !found {
		*dbSub = append(*dbSub, device)
	}

	if err := s.subscriptionRepository.Save(callerID, dbSub); err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) Delete(ctx context.Context, uuid string) error {
	// Fetch the callerID from the given context.
	callerID, ok := ctx.Value("nickname").(string)
	if !ok {
		return fmt.Errorf("could not decode the caller's ID")
	}

	dbSub, err := s.subscriptionRepository.GetByUserID(callerID)
	if err != nil {
		return err
	}

	// Prepare a new array for remaining devices.
	var newDevices []models.Device

	// Loop through fetched devices, skip the requested one as well as the blank ones.
	for _, dev := range *dbSub {
		if dev.UUID == uuid {
			// do not include this device anymore
			continue
		}

		if reflect.DeepEqual(dev, (models.Device{})) {
			continue
		}

		if dev.UUID == "" {
			continue
		}

		// append non matching device
		newDevices = append(newDevices, dev)
	}

	if err := s.subscriptionRepository.Save(callerID, &newDevices); err != nil {
		return err
	}

	return nil
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
