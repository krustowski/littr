package posts

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	chi "github.com/go-chi/chi/v5"
	app "github.com/maxence-charriere/go-app/v9/pkg/app"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/live"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/models"
)

const (
	PKG_NAME = "posts"
)

// getPosts fetches posts, page spicified by a header.
//
// @Summary      Get posts
// @Description  get posts
// @Tags         posts
// @Produce      json
// @Param    	 X-Page-No header string true "page number"
// @Param    	 X-Hide-Replies header string false "hide replies bool"
// @Success      200  {object}  common.APIResponse{data=posts.getPosts.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts [get]
func getPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	type responseData struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
		Count int                    `json:"count"`
	}

	// skip blank callerID
	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pageNo := 0

	pageNoString := r.Header.Get(common.HDR_PAGE_NO)
	if page, err := strconv.Atoi(pageNoString); err != nil {
		//l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		//return
	} else {
		pageNo = page
	}

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

// addNewPost adds new post
//
// @Summary      Add new post
// @Description  add new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param    	 request body models.Post true "new post struct in request body"
// @Success      201  {object}  common.APIResponse{data=posts.addNewPost.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts [post]
func addNewPost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

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
		devs, _ := db.GetOne(db.SubscriptionCache, receiverName, []models.Device{})
		_, found := db.GetOne(db.UserCache, receiverName, models.User{})
		if !found {
			continue
		}

		// do not notify the same person --- OK condition
		if receiverName == l.CallerID() {
			continue
		}

		// do not notify user --- notifications disabled --- OK condition
		if len(devs) == 0 {
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
			Devices:  &devs,
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

// updatePostStarCount increases the star count for the given post.
//
// @Summary      Update post's star count
// @Description  update the star count
// @Tags         posts
// @Produce      json
// @Param        postID path string true "post's ID to update"
// @Success      200  {object}  common.APIResponse{data=posts.updatePostStarCount.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID}/star [patch]
func updatePostStarCount(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

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

// updatePost updates the specified post.
//
// @Summary      Update specified post
// @Description  update specified post
// @Deprecated
// @Tags         posts
// @Accept       json
// @Produce      json
// @Param    	 request body models.Post true "post to update in request body"
// @Param        postID path string true "post's ID to update"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID} [put]
func updatePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	if l.CallerID() == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var post models.Post

	// decode the request data
	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	// check if suck post even exists
	if _, found := db.GetOne(db.FlowCache, post.ID, models.Post{}); !found {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// check for the post update forgery
	if post.Nickname != l.CallerID() {
		l.Msg(common.ERR_POST_UPDATE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// save the updated post back (whole decoded struct to override the existing one --- very dangerous and nasty)
	// one could easily change post's author to oneself and mutilate the post afterwards
	if saved := db.SetOne(db.FlowCache, post.ID, post); !saved {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	l.Msg("ok, post updated").Status(http.StatusOK).Log().Payload(nil).Write(w)
	return
}

// deletePost removes specified post.
//
// @Summary      Delete specified post
// @Description  delete specified post
// @Tags         posts
// @Produce      json
// @Param        postID path string true "post's ID to update"
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID} [delete]
func deletePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

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

// getSinglePost fetch specified post and its interaction.
//
// @Summary      Get single post
// @Description  get single post
// @Tags         posts
// @Produce      json
// @Param    	 X-Hide-Replies header string false "hide replies"
// @Param    	 X-Page-No header string true "page number"
// @Param        postID path string true "post's ID to update"
// @Success      200  {object}  common.APIResponse{data=posts.getSinglePost.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID} [get]
func getSinglePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

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

// fetchHashtaggedPosts
//
// @Summary      Get hashtagged post list
// @Description  get hashtagged post list
// @Tags         posts
// @Produce      json
// @Param    	 X-Hide-Replies header string false "hide replies"
// @Param    	 X-Page-No header string true "page number"
// @Param        hashtag path string true "hashtag string"
// @Success      200  {object}  common.APIResponse{data=posts.fetchHashtaggedPosts.responseData}
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/hashtags/{hashtag} [get]
func fetchHashtaggedPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

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
