package posts

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"
	app "github.com/maxence-charriere/go-app/v10/pkg/app"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/models"
)

const (
	LOGGER_WORKER_NAME = "postController"
)

// Structure contents definition for the controller.
type PostController struct {
	postService models.PostServiceInterface
}

// NewPostController return a pointer to the new controller instance, that has to be populated with the Post service.
func NewPostController(
	postService models.PostServiceInterface,
) *PostController {

	if postService == nil {
		return nil
	}

	return &PostController{
		postService: postService,
	}
}

// GetAll fetches a list of posts, a page number is specified by the X-Page-No header.
//
//	@Summary		Get posts
//	@Description		This function call retrieves a page of posts. The page number is to be specified using the `X-Page-No` header (default is 0 = latest).
//	@Tags			posts
//	@Produce		json
//	@Param			X-Page-No		header		integer	false		"A page number (default is 0)."
//	@Param			X-Hide-Replies		header		bool	false		"An optional boolean to show only root posts without any reply (default is false)."
//	@Success		200				{object}	common.APIResponse{data=posts.GetAll.responseData}	"Paginated list of posts."
//	@Failure		400				{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		401				{object}	common.APIResponse{data=models.Stub}			"User unauthorized."
//	@Failure		429				{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		500				{object}	common.APIResponse
//	@Router			/posts [get]
func (c *PostController) GetAll(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
		Count int                    `json:"count"`
	}

	// Skip blank callerID.
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pageNo := 0

	// Parse the X-Page-No header.
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		pageNo = 0
	}

	// Parse the X-Hide-Replies header.
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	opts := pages.PageOptions{
		CallerID: l.CallerID(),
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies: hideReplies,
			Plain:       hideReplies == false,
		},
	}

	// fetch page according to the logged user
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || pagePtrs.Users == nil || (*pagePtrs.Posts) == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, l.CallerID(), models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		(*pagePtrs.Users)[l.CallerID()] = caller
	}

	// compose the payload
	pl := &responseData{
		Posts: *pagePtrs.Posts,
		Users: *common.FlushUserData(pagePtrs.Users, l.CallerID()),
		Key:   l.CallerID(),
		Count: pages.PAGE_SIZE,
	}

	l.Msg("ok, dumping posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// Create handles a new post creation request to the post service, which adds the post to the database.
//
//	@Summary		Add new post
//	@Description		This function call is to be used to create a new post.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body	posts.PostCreateRequest			true				"Post body."
//	@Success		201		{object}	common.APIResponse{data=posts.Create.responseData}	"New post has been added to the database and published."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}			"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}			"Forbidden action occurred."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}			"Internal server problem occurred while processing the request."
//	@Router			/posts [post]
func (c *PostController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
	}

	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var post models.Post

	// decode received data
	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if post.Content == "" && post.Figure == "" && post.Data == nil {
		l.Msg(common.ERR_POST_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// patch wrongly loaded user data from LocalStorage on FE
	if post.Nickname == "" {
		post.Nickname = l.CallerID()
	}

	// check the post forgery possibility
	if l.CallerID() != post.Nickname {
		l.Msg(common.ERR_POSTER_INVALID).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// prepare an ID for the post
	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	post.ID = key
	post.Timestamp = timestamp

	// Compose a payload for the image processing.
	imagePayload := &image.ImageProcessPayload{
		ImageByteData: &post.Data,
		ImageFileName: post.Figure,
		ImageBaseName: post.ID,
	}

	// Uploaded figure handling.
	if post.Data != nil && post.Figure != "" {
		var err error
		imgReference, err := image.ProcessImageBytes(imagePayload)
		if err != nil {
			l.Status(common.DecideStatusFromError(err)).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// Ensure the image reference pointer is a valid one.
		if imgReference != nil {
			post.Figure = *imgReference
		}

		post.Data = make([]byte, 0)
	}

	// save the post by its key
	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// notify all to-be-notifiedees
	//
	// find matches using regexp compiling to '@username' matrix
	re := regexp.MustCompile(`@(?P<text>\w+)`)
	matches := re.FindAllStringSubmatch(post.Content, -1)

	// we deal with a 2D array here
	for _, match := range matches {
		receiverName := match[1]

		// fetch related data from cachces
		receiver, found := db.GetOne(db.UserCache, receiverName, models.User{})
		if !found {
			continue
		}

		// do not notify the same person --- OK condition
		if receiverName == l.CallerID() {
			continue
		}

		// do not notify user --- notifications disabled --- OK condition
		if len(receiver.Devices) == 0 {
			continue
		}

		// compose the body of this notification
		body, err := json.Marshal(app.Notification{
			Title: "littr mention",
			Icon:  "/web/apple-touch-icon.png",
			Body:  l.CallerID() + " mentioned you in their post",
			Path:  "/flow/posts/" + post.ID,
		})
		if err != nil {
			l.Msg(common.ERR_PUSH_BODY_COMPOSE_FAIL).Status(http.StatusInternalServerError).Log()
			continue
		}

		opts := &push.NotificationOpts{
			Receiver: receiverName,
			Devices:  &receiver.Devices,
			Body:     &body,
			//Logger:   l,
		}

		// send the webpush notification(s)
		push.SendNotificationToDevices(opts)
	}

	// broadcast a new post to live subscribers
	live.BroadcastMessage(live.EventPayload{Data: "post," + post.Nickname, Type: "message"})

	// prepare the payload
	posts := make(map[string]models.Post)
	posts[key] = post

	pl := &responseData{
		Posts: posts,
	}

	l.Msg("ok, adding new post").Status(http.StatusCreated).Log().Payload(pl).Write(w)
	return
}

// UpdateReactions increases the star count for the given post.
//
//	@Summary		Update post's star count
//	@Description		This function call is used to increase the post reactions counter value. It add one (1) reaction to the current state of a post.
//	@Tags			posts
//	@Produce		json
//	@Param			postID	path			string		true						"Post ID to update."
//	@Success		200		{object}	common.APIResponse{data=posts.UpdateReactions.responseData}	"Counter value has been increased successfully."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}				"Invalid data input."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}				"Forbidden action happened (e.g. caller tried to increase the counter value of a post of theirs)."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}				"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}				"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}				"Internal server problem occurred while processing the request."
//	@Router			/posts/{postID}/star [patch]
func (c *PostController) UpdateReactions(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
	}

	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	postID := chi.URLParam(r, "postID")
	if postID == "" {
		l.Msg(common.ERR_POSTID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the post using postID
	post, found := db.GetOne(db.FlowCache, postID, models.Post{})
	if !found {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check if there is a selfrater
	if post.Nickname == l.CallerID() {
		l.Msg(common.ERR_POST_SELF_RATE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// increment the star count by one (1)
	post.ReactionCount++

	// resave the post back to databse
	if saved := db.SetOne(db.FlowCache, postID, post); !saved {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare payload
	posts := make(map[string]models.Post)
	posts[postID] = post

	pl := &responseData{
		Posts: posts,
	}

	l.Msg("ok, star count incremented").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// Delete removes the specified post.
//
//	@Summary		Delete specified post
//	@Description		This function call ensures that the specified post is purged from the database. Associated items like figures are deleted as well.
//	@Tags			posts
//	@Produce		json
//	@Param			postID		path		string		true			"Post ID to delete."
//	@Success		200		{object}	common.APIResponse{data=models.Stub}	"Specified post has been deleted."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}	"Invalid data input."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}	"Forbidden action occurred (e.g. caller tried to delete a foreign post)."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}	"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}	"Too many requests, try again later."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}	"Internal server problem occurred while processing the request."
//	@Router			/posts/{postID} [delete]
func (c *PostController) Delete(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	// skip blank callerID
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	postID := chi.URLParam(r, "postID")
	if postID == "" {
		l.Msg(common.ERR_POSTID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the post using postID
	post, found := db.GetOne(db.FlowCache, postID, models.Post{})
	if !found {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the possible post forgery
	if post.Nickname != l.CallerID() {
		l.Msg(common.ERR_POST_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// delete such post
	if deleted := db.DeleteOne(db.FlowCache, postID); !deleted {
		l.Msg(common.ERR_POST_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// delete associated image and its thumbnail
	if post.Figure != "" {
		err := os.Remove("/opt/pix/thumb_" + post.Figure)
		if err != nil {
			l.Msg(common.ERR_POST_DELETE_THUMB).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		err = os.Remove("/opt/pix/" + post.Figure)
		if err != nil {
			l.Msg(common.ERR_POST_DELETE_FULLIMG).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}
	}

	l.Msg("ok, post removed").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// GetByID fetches specified post and its interaction.
//
//	@Summary		Get single post
//	@Description		This function call enables one to fetch a single post by its ID with all replies associated.
//	@Tags			posts
//	@Produce		json
//	@Param			X-Hide-Replies		header		bool		false						"Optional parameter to hide all replies (default is false)."
//	@Param			X-Page-No		header		integer		false						"Page number (default is 0)."
//	@Param			postID			path		string		true						"Post ID to fetch."
//	@Success		200			{object}	common.APIResponse{data=posts.GetByID.responseData}		"Data fetched successfully."
//	@Failure		400			{object}	common.APIResponse{data=models.Stub}				"Invalid input data."
//	@Failure		401			{object}	common.APIResponse{data=models.Stub}				"User unauthorized."
//	@Failure		429			{object}	common.APIResponse{data=models.Stub}				"Too many requests, try again later."
//	@Failure		500			{object}	common.APIResponse{data=models.Stub}				"Internal server problem occurred while processing the request."
//	@Router			/posts/{postID} [get]
func (c *PostController) GetByID(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
	}

	// skip blank callerID
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	postID := chi.URLParam(r, "postID")
	if postID == "" {
		l.Msg(common.ERR_POSTID_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the required X-Page-No header
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// fetch the optional X-Hide-Replies header
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	opts := pages.PageOptions{
		CallerID: l.CallerID(),
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies:  hideReplies,
			Plain:        hideReplies == false,
			SinglePost:   true,
			SinglePostID: postID,
		},
	}

	// fetch page according to the logged user
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || pagePtrs.Users == nil || (*pagePtrs.Posts) == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, l.CallerID(), models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		(*pagePtrs.Users)[l.CallerID()] = caller
	}

	// prepare the payload
	pl := &responseData{
		Posts: *pagePtrs.Posts,
		Users: *common.FlushUserData(pagePtrs.Users, l.CallerID()),
		Key:   l.CallerID(),
	}

	l.Msg("ok, dumping single post and its interactions").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}

// GetByHashtag fetches all posts tagged with the specified hashtag.
//
//	@Summary		Get hashtagged post list
//	@Description		This function call fetches all posts tagged with specified hashtag phrase (`#phrase` => `phrase`).
//	@Tags			posts
//	@Produce		json
//	@Param			X-Hide-Replies		header		bool		false							"hide replies"
//	@Param			X-Page-No		header		integer		false							"page number"
//	@Param			hashtag			path		string		true							"hashtag string"
//	@Success		200			{object}	common.APIResponse{data=posts.GetByHashtag.responseData}		"Data fetched successfully."
//	@Failure		400			{object}	common.APIResponse{data=models.Stub}					"Invalid input data."
//	@Failure		401			{object}	common.APIResponse{data=models.Stub}					"User unauthorized."
//	@Failure		429			{object}	common.APIResponse{data=models.Stub}					"Too many requests, try again later."
//	@Failure		500			{object}	common.APIResponse{data=models.Stub}
//	@Router			/posts/hashtags/{hashtag} [get]
func (c *PostController) GetByHashtag(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, LOGGER_WORKER_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
	}

	// skip blank callerID
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// take the param from path
	hashtag := chi.URLParam(r, "hashtag")
	if hashtag == "" {
		l.Msg(common.ERR_HASHTAG_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// fetch the required X-Page-No header
	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	pageNo, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// fetch the optional X-Hide-Replies header
	hideReplies, err := strconv.ParseBool(r.Header.Get(common.HDR_HIDE_REPLIES))
	if err != nil {
		//l.Msg(common.ERR_HIDE_REPLIES_INVALID).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
		hideReplies = false
	}

	opts := pages.PageOptions{
		CallerID: l.CallerID(),
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies: hideReplies,
			Plain:       hideReplies == false,
			Hashtag:     hashtag,
		},
	}

	// fetch page according to the logged user
	pagePtrs := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || pagePtrs.Users == nil || (*pagePtrs.Posts) == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, l.CallerID(), models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		(*pagePtrs.Users)[l.CallerID()] = caller
	}

	// prepare the payload
	pl := &responseData{
		Posts: *pagePtrs.Posts,
		Users: *common.FlushUserData(pagePtrs.Users, l.CallerID()),
		Key:   l.CallerID(),
	}

	l.Msg("ok, dumping hastagged posts and their parent posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
	return
}
