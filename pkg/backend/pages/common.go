// universal cache pagination tooling
package pages

import (
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/models"
)

const PAGE_SIZE int = 25

// DTO for GetOnePage input aggregation
type PageOptions struct {
	// common options
	Caller   models.User
	CallerID string          `json:"caller_id"`
	PageNo   int             `json:"page_no"`
	FlowList map[string]bool `json:"folow_list"`

	// data compartments' specifications
	Flow  FlowOptions `json:"flow_options"`
	Polls PollOptions `json:"poll_options"`
	Users UserOptions `json:"user_options"`
}

// flow subviews' options
type FlowOptions struct {
	Plain        bool   `json:"plain"`
	SinglePost   bool   `json:"single_post"`
	UserFlow     bool   `json:"user_flow"`
	SinglePostID string `json:"single_post_id"`
	UserFlowNick string `json:"user_Flow_nick"`
	Hashtag      string `json:"hashtag"`
	HideReplies  bool   `json:"hide_replies"`
}

// polls subviews' options
type PollOptions struct {
	Plain        bool   `json:"plain"`
	SinglePoll   bool   `json:"single_poll"`
	SinglePollID string `json:"single_poll_id"`
}

// users subviews' options
type UserOptions struct {
	Plain        bool   `json:"plain"`
	SingleUser   bool   `json:"single_user"`
	SingleUserID string `json:"single_user_id"`
}

// DTO for GetOnePage pointer output aggregation
type PagePointers struct {
	Polls *map[string]models.Poll
	Posts *map[string]models.Post
	Users *map[string]models.User
}

// fillDataMaps is a function, that prepares raw maps of all (related) items for further processing according to input options
func fillDataMaps(opts PageOptions) *PagePointers {
	// prepare data maps for a flow page
	if opts.Flow != (FlowOptions{}) {
		posts, _ := db.GetAll(db.FlowCache, models.Post{})
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &PagePointers{Posts: &posts, Users: &users}
	}

	// prepare data map for a polls page
	if opts.Polls != (PollOptions{}) {
		polls, _ := db.GetAll(db.PollCache, models.Poll{})
		// users are not needed necessarily there for now...
		//users, _ := db.GetAll(db.UserCache, models.User{})

		return &PagePointers{Polls: &polls}
	}

	// prepare data map for a users page
	if opts.Users != (UserOptions{}) {
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &PagePointers{Users: &users}
	}

	return nil
}

func GetOnePage(opts PageOptions) (ptrs PagePointers) {
	var ok bool

	// validate the callerID is a legitimate user's ID
	if opts.Caller, ok = db.GetOne(db.UserCache, opts.CallerID, models.User{}); !ok {
		// unregistred caller
		return ptrs
	}

	// pointer to maps of all items (based on and related to the opts input)
	ptrMaps := fillDataMaps(opts)
	if ptrMaps == nil {
		// invalid input options = resulted in empty maps only
		return ptrs
	}

	if opts.Flow != (FlowOptions{}) {
		return onePagePosts(opts, ptrMaps)
	}

	if opts.Polls != (PollOptions{}) {
		return onePagePolls(opts, ptrMaps)
	}

	if opts.Users != (UserOptions{}) {
		return onePageUsers(opts, ptrMaps)
	}

	return ptrs
}
