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
	Flow  flowOptions `json:"flow_options"`
	Polls pollOptions `json:"poll_options"`
	Users userOptions `json:"user_options"`
}

// flow subviews' options
type flowOptions struct {
	SinglePost   bool   `json:"single_post"`
	UserFlow     bool   `json:"user_flow"`
	SinglePostID string `json:"single_post_id"`
	UserFlowNick string `json:"user_Flow_nick"`
	Hashtag      string `json:"hashtag" default:""`
	HideReplies  bool   `json:"hide_replies"`
}

// polls subviews' options
type pollOptions struct {
	SinglePoll   bool   `json:"single_poll"`
	SinglePollID string `json:"single_poll_id"`
}

// users subviews' options
type userOptions struct {
	SingleUser   bool   `json:"single_user"`
	SingleUserID string `json:"single_user_id"`
}

// DTO for fillMaps output aggregation
type maps struct {
	Polls map[string]models.Poll
	Posts map[string]models.Post
	Users map[string]models.User
}

// DTO for GetOnePage pointer output aggregation
type PagePointers struct {
	PtrPolls *map[string]models.Poll
	PtrPosts *map[string]models.Post
	PtrUsers *map[string]models.User
}

// fillDataMaps is a function, that prepares raw maps of all (related) items for further processing according to input options
func fillDataMaps(opts PageOptions) *maps {
	// prepare data maps for a flow page
	if opts.Flow != (flowOptions{}) {
		posts, _ := db.GetAll(db.FlowCache, models.Post{})
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Posts: posts, Users: users}
	}

	// prepare data map for a polls page
	if opts.Polls != (pollOptions{}) {
		polls, _ := db.GetAll(db.PollCache, models.Poll{})
		// users are not needed necessarily there for now...
		//users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Polls: polls}
	}

	// prepare data map for a users page
	if opts.Users != (userOptions{}) {
		users, _ := db.GetAll(db.UserCache, models.User{})

		return &maps{Users: users}
	}

	return nil
}

func GetOnePage(opts PageOptions) PagePointers {
	// validate the callerID is a legitimate user's ID
	if opts.Caller, ok := db.GetOne(db.UserCache, opts.CallerID, models.User{}); !ok {
		// unregistred caller
		return nil
	}

	// pointer to maps of all items (based on and related to the opts input)
	ptrMaps := fillDataMaps(opts)
	if ptrMaps == nil {
		// invalid input options = resulted in empty maps only
		return ptrMaps
	}

	if opts.Flow != nil {
		return onePageFlow(opts, ptrMaps)
	}

	if opts.Polls != nil {
		// NYI
		//return onePagePolls(opts, ptrMaps)
	}

	if opts.Users != nil {
		// NYI
		//return onePageUsers(opts, ptrMaps)
	}

	return nil
}
