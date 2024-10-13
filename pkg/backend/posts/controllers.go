package posts

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	sse "github.com/alexandrevicenzi/go-sse"
	chi "github.com/go-chi/chi/v5"
	app "github.com/maxence-charriere/go-app/v9/pkg/app"

	"go.vxn.dev/littr/pkg/backend/common"
	"go.vxn.dev/littr/pkg/backend/db"
	"go.vxn.dev/littr/pkg/backend/image"
	"go.vxn.dev/littr/pkg/backend/pages"
	"go.vxn.dev/littr/pkg/backend/push"
	"go.vxn.dev/littr/pkg/models"
)

const (
	HDR_PAGE_NO = "X-Page-No"
	PKG_NAME    = "posts"
)

// getPosts fetches posts, page spicified by a header.
//
// @Summary      Get posts
// @Description  get posts
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/ [get]
func getPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pageNo := 0

	pageNoString := r.Header.Get(HDR_PAGE_NO)
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	pageNo = page

	if callerID == "" {
		l.Msg(common.ERR_CALLER_BLANK).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	hideReplies, err := strconv.ParseBool(r.Header.Get("X-Hide-Replies"))
	if err != nil {
		/*resp.Message = "invalid X-Hide-Replies"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return*/
		hideReplies = false
	}

	opts := pages.PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,

		Flow: pages.FlowOptions{
			HideReplies: hideReplies,
			Plain:       hideReplies == false,
		},
	}

	// fetch page according to the logged user
	pagePtrs := pages.GetOnePage(opts)
	//pExport, uExport := pages.GetOnePage(opts)
	if pagePtrs == (pages.PagePointers{}) || pagePtrs.Posts == nil || pagePtrs.Users == nil || (*pagePtrs.Posts) == nil || (*pagePtrs.Users) == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		(*pagePtrs.Users)[callerID] = caller
	}

	for key, user := range *pagePtrs.Users {
		user.Passphrase = ""
		user.PassphraseHex = ""
		user.Email = ""

		if user.Nickname != callerID {
			user.FlowList = nil
			user.ShadeList = nil
			user.RequestList = nil
		}

		(*pagePtrs.Users)[key] = user
	}

	pl := struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
		Count int                    `json:"count"`
	}{
		Posts: *pagePtrs.Posts,
		Users: *pagePtrs.Users,
		Key:   callerID,
		Count: pages.PAGE_SIZE,
	}

	l.Msg("ok, dumping posts").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// addNewPost adds new post
//
// @Summary      Add new post
// @Description  add new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      201  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/ [post]
func addNewPost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	// get caller's nickname from context
	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
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
		post.Nickname = callerID
	}

	// check the post forgery possibility
	if callerID != post.Nickname {
		l.Msg(common.ERR_POSTER_INVALID).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	post.ID = key
	post.Timestamp = timestamp

	var imgReference string

	// uploaded figure handling
	if post.Data != nil && post.Figure != "" {
		if err, code := image.ProcessBytes(&post, &imgReference); err != nil {
			l.Status(code).Error(err).Log().Payload(nil).Write(w)
			return
		}
	}

	// save the post by its key
	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// notify all to-be-notifiedees
	//
	// find matches using regext compiling to '@username' matrix
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
		if receiverName == callerID {
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
			Body:  callerID + " mentioned you in their post",
			Path:  "/flow/post/" + post.ID,
		})
		if err != nil {
			l.Msg(common.ERR_PUSH_BODY_COMPOSE_FAIL).Status(http.StatusInternalServerError).Log()
			continue
		}

		push.SendNotificationToDevices(receiverName, devs, body, l)
	}

	// broadcast a new post to live subscribers
	if Streamer != nil {
		Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("post,"+post.Nickname))
	}

	posts := make(map[string]models.Post)
	posts[key] = post

	pl := struct {
		Posts map[string]models.Post `json:"posts"`
	}{
		Posts: posts,
	}

	l.Msg("ok, adding new post").Status(http.StatusCreated).Log().Payload(&pl).Write(w)
	return
}

// updatePostStarCount increases the star count for the given post.
//
// @Summary      Update post's star count
// @Description  update the star count
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID}/star [patch]
func updatePostStarCount(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	var found bool

	if post, found = db.GetOne(db.FlowCache, key, models.Post{}); !found {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if post.Nickname == callerID {
		l.Msg(common.ERR_POST_SELF_RATE).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	// increment the star count by 1
	post.ReactionCount++

	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		l.Msg(common.ERR_POST_SAVE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	posts := make(map[string]models.Post)
	posts[key] = post

	pl := struct {
		Posts map[string]models.Post `json:"posts"`
	}{
		Posts: posts,
	}

	l.Msg("ok, star count incremented").Status(http.StatusOK).Log().Payload(&pl).Write(w)
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
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID} [put]
func updatePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if _, found := db.GetOne(db.FlowCache, key, models.Post{}); !found {
		l.Msg(common.ERR_POST_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	if post.Nickname != callerID {
		l.Msg(common.ERR_POST_UPDATE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if saved := db.SetOne(db.FlowCache, key, post); !saved {
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
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Failure      403  {object}  common.APIResponse
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/{postID} [delete]
func deletePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// remove a post
	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		l.Msg(common.ERR_INPUT_DATA_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if post.Nickname != callerID {
		l.Msg(common.ERR_POST_DELETE_FOREIGN).Status(http.StatusForbidden).Log().Payload(nil).Write(w)
		return
	}

	if deleted := db.DeleteOne(db.FlowCache, key); !deleted {
		l.Msg(common.ERR_POST_DELETE_FAIL).Status(http.StatusInternalServerError).Log().Payload(nil).Write(w)
		return
	}

	// delete associated image and thumbnail
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
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Router       /posts/{postID} [get]
func getSinglePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	//user, _ := getOne(UserCache, callerID, models.User{})

	// parse the URI's path
	// check if diff page has been requested
	postID := chi.URLParam(r, "postID")

	pageNo := 0

	pageNoString := r.Header.Get(HDR_PAGE_NO)
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	pageNo = page

	/*flowList := user.FlowList
	if flowList == nil {
		flowList = make(map[string]bool)
	}*/

	opts := PageOptions{
		SinglePost:   true,
		SinglePostID: postID,
		CallerID:     callerID,
		PageNo:       pageNo,
		//FlowList:   flowList,
	}

	// fetch page according to the logged user
	pExport, uExport := GetOnePage(opts)
	if pExport == nil || uExport == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// flush email addresses
	uExport = *common.FlushUserData(&uExport, callerID)

	/*for key, user := range uExport {
		if key == callerID {
			continue
		}
		user.Email = ""
		uExport[key] = user
	}*/

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	pl := struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
	}{
		Posts: pExport,
		Users: uExport,
		Key:   callerID,
	}

	l.Msg("ok, dumping single post and its interactions").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}

// fetchHashtaggedPosts
//
// @Summary      Get hashtagged post list
// @Description  get hashtagged post list
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.APIResponse
// @Failure      400  {object}  common.APIResponse
// @Router       /posts/hashtag/{hashtag} [get]
func fetchHashtaggedPosts(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, PKG_NAME)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// parse the URI's path
	// check if diff page has been requested
	hashtag := chi.URLParam(r, "hashtag")

	pageNoString := r.Header.Get(HDR_PAGE_NO)
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		l.Msg(common.ERR_PAGENO_INCORRECT).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
		return
	}

	opts := PageOptions{
		Hashtag:  hashtag,
		CallerID: callerID,
		PageNo:   page,
	}

	// fetch page according to the logged user
	pExport, uExport := GetOnePage(opts)
	if pExport == nil || uExport == nil {
		l.Msg(common.ERR_PAGE_EXPORT_NIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// flush email addresses
	uExport = *common.FlushUserData(&uExport, callerID)

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		l.Msg(common.ERR_CALLER_NOT_FOUND).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	pl := struct {
		Posts map[string]models.Post `json:"posts"`
		Users map[string]models.User `json:"users"`
		Key   string                 `json:"key"`
	}{
		Posts: pExport,
		Users: uExport,
		Key:   callerID,
	}

	l.Msg("ok, dumping hastagged posts and their parent posts").Status(http.StatusOK).Log().Payload(&pl).Write(w)
	return
}
