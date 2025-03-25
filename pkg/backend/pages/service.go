package pages

import (
	"context"
	"errors"

	"go.vxn.dev/littr/pkg/backend/common"
)

var (
	errCorruptedData            = errors.New("data corrupted")
	errEmptyDataOutput          = errors.New("no data to send")
	errInvalidPayload           = errors.New("invalid payload was received")
	errNotImplementedYet        = errors.New("method not implemented yet")
	errUnknownOriginNotAccepted = errors.New("request origin not specified")
)

type pagingService struct {
	//
}

func NewPagingService() *pagingService {
	return &pagingService{}
}

func (s *pagingService) GetOne(ctx context.Context, options interface{}, data ...interface{}) (interface{}, error) {
	callerID := common.GetCallerID(ctx)

	opts, ok := options.(*PageOptions)
	if !ok {
		return nil, errInvalidPayload
	}

	opts.CallerID = callerID

	if opts.Flow != (FlowOptions{}) {
		return onePagePosts(opts, data), nil
	}

	if opts.Polls != (PollOptions{}) {
		return onePagePolls(opts, data), nil
	}

	if opts.Users != (UserOptions{}) {
		return onePageUsers(opts, data), nil
	}
	//
	//
	//

	/*var data []interface{}

	for _, cache := range opts.Caches {
		switch cache.GetName() {
		case "FlowCache":
			genericMap, _ := cache.Range()

			m := make(map[string]models.Post)

			for key, valI := range *genericMap {
				val, ok := valI.(models.Post)
				if !ok {
					continue
				}

				m[key] = val
			}

			data = append(data, &m)

		case "UserCache":
			genericMap, _ := cache.Range()

			m := make(map[string]models.User)

			for key, valI := range *genericMap {
				val, ok := valI.(models.User)
				if !ok {
					continue
				}

				m[key] = val
			}

			data = append(data, &m)
		}
	}

	if data == nil || len(data) == 0 {
		// invalid input options = resulted in empty maps only
		return nil, errEmptyDataOutput
	}

	if opts.Flow != (FlowOptions{}) {
		return onePagePosts(opts, data), nil
	}

	/*if opts.Polls != (PollOptions{}) {
		return onePagePolls(opts, data), nil
	}

	if opts.Users != (UserOptions{}) {
		return onePageUsers(opts, data), nil
	}*/

	return nil, nil
}

func (s *pagingService) GetMany(ctx context.Context, options interface{}) (interface{}, error) {
	return nil, errNotImplementedYet
}
