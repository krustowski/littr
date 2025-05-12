package posts

import (
	"net/http"
	"os"
	"strconv"

	chi "github.com/go-chi/chi/v5"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/models"
)

const (
	loggerWorkerName = "postController"
)

// Structure contents definition for the controller.
type PostController struct {
	postService models.PostServiceInterface
	userService models.UserServiceInterface
}

// NewPostController return a pointer to the new controller instance, that has to be populated with the Post service.
func NewPostController(
	postService models.PostServiceInterface,
	userService models.UserServiceInterface,
) *PostController {

	if postService == nil || userService == nil {
		return nil
	}

	return &PostController{
		postService: postService,
		userService: userService,
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
	l := common.NewLogger(r, loggerWorkerName)

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

	opts := &PostPagingRequest{
		HideReplies: hideReplies,
		PageNo:      pageNo,
		PagingSize:  25,
	}

	posts, users, err := c.postService.FindAll(r.Context(), opts)
	if err != nil {
		l.Error(err).Log()
		return
	}

	_, err = c.userService.FindByID(r.Context(), l.CallerID())
	if err != nil {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Error(err).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		//
	}

	// compose the payload
	pl := &responseData{
		Posts: *posts,
		Users: *users,
		Key:   l.CallerID(),
		Count: pages.PAGE_SIZE,
	}

	l.Msg("ok, dumping posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
}

// Create handles a new post creation request to the post service, which adds the post to the database.
//
//	@Summary		Add new post
//	@Description		This function call is to be used to create a new post.
//	@Tags			posts
//	@Accept			json
//	@Produce		json
//	@Param			request	body	posts.PostCreateRequest			true				"Post body."
//	@Success		201		{object}	common.APIResponse{data=models.Post}			"New post has been added to the database and published."
//	@Failure		400		{object}	common.APIResponse{data=models.Stub}			"Invalid input data."
//	@Failure		401		{object}	common.APIResponse{data=models.Stub}			"User unauthorized."
//	@Failure		429		{object}	common.APIResponse{data=models.Stub}			"Too many requests, try again later."
//	@Failure		403		{object}	common.APIResponse{data=models.Stub}			"Forbidden action occurred."
//	@Failure		500		{object}	common.APIResponse{data=models.Stub}			"Internal server problem occurred while processing the request."
//	@Router			/posts [post]
func (c *PostController) Create(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, loggerWorkerName)

	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	if err := c.postService.Create(r.Context(), &post); err != nil {
		l.Msg("postService: ").Error(err).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, adding new post").Status(http.StatusCreated).Log().Payload(post).Write(w)
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
	l := common.NewLogger(r, loggerWorkerName)

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

	post, _, err := c.postService.FindByID(r.Context(), postID)
	if err != nil {
		l.Error(err).Status(http.StatusInternalServerError).Log().Write(w)
		return
	}

	// check if there is a selfrater
	if post.Nickname == l.CallerID() {
		l.Msg(common.ERR_POST_SELF_RATE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// increment the star count by one (1)
	post.ReactionCount++

	if err := c.postService.Update(r.Context(), post); err != nil {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// prepare payload
	posts := make(map[string]models.Post)
	posts[postID] = *post

	pl := &responseData{
		Posts: posts,
	}

	l.Msg("ok, star count incremented").Status(http.StatusOK).Log().Payload(pl).Write(w)
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
	l := common.NewLogger(r, loggerWorkerName)

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

	post, _, err := c.postService.FindByID(r.Context(), postID)
	if err != nil {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the possible post forgery
	if post.Nickname != l.CallerID() {
		l.Msg(common.ERR_POST_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if err := c.postService.Delete(r.Context(), postID); err != nil {
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
	l := common.NewLogger(r, loggerWorkerName)

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

	opts := &PostPagingRequest{
		HideReplies:  hideReplies,
		PageNo:       pageNo,
		PagingSize:   25,
		SinglePost:   true,
		SinglePostID: postID,
	}

	posts, users, err := c.postService.FindAll(r.Context(), opts)
	if err != nil {
		l.Error(err).Log()
		return
	}

	// prepare the payload
	pl := &responseData{
		Posts: *posts,
		Users: *common.FlushUserData(users, l.CallerID()),
		Key:   l.CallerID(),
	}

	l.Msg("ok, dumping single post and its interactions").Status(http.StatusOK).Log().Payload(pl).Write(w)
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
	l := common.NewLogger(r, loggerWorkerName)

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

	opts := &PostPagingRequest{
		HideReplies: hideReplies,
		PageNo:      pageNo,
		PagingSize:  25,
		Hashtag:     hashtag,
	}

	posts, users, err := c.postService.FindAll(r.Context(), opts)
	if err != nil {
		l.Error(err).Log()
		return
	}

	// prepare the payload
	pl := &responseData{
		Posts: *posts,
		Users: *common.FlushUserData(users, l.CallerID()),
		Key:   l.CallerID(),
	}

	l.Msg("ok, dumping hastagged posts and their parent posts").Status(http.StatusOK).Log().Payload(pl).Write(w)
}
