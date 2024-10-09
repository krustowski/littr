package posts

import (
	"encoding/json"
	pic "image"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
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
	pkgName = "posts"
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
	l := common.NewLogger(r, pkgName)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	pageNo := 0

	pageNoString := r.Header.Get("X-Page-No")
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
// @Failure      500  {object}  common.APIResponse
// @Router       /posts/ [post]
func addNewPost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, pkgName)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// post a new post
	var post models.Post

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
		l.Msg(common.ERR_POSTER_INVALID).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	post.ID = key
	post.Timestamp = timestamp

	// uploadedFigure handling
	if post.Data != nil && post.Figure != "" {
		var newBytes *[]byte
		var err error
		var img *pic.Image
		var format string

		fileExplode := strings.Split(post.Figure, ".")
		extension := fileExplode[len(fileExplode)-1]

		// decode image from []byte stream
		img, format, err = image.DecodeImage(&post.Data, extension)
		if err != nil {
			l.Msg(common.ERR_IMG_DECODE_FAIL).Status(http.StatusBadRequest).Error(err).Log().Payload(nil).Write(w)
			return
		}

		switch extension {
		case "png", "jpg", "jpeg":
			// fix the image orientation for decoded image
			img, err = image.FixOrientation(img, &post.Data)
			if err != nil {
				l.Msg(common.ERR_IMG_ORIENTATION_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
				return
			}

			// re-encode the image to flush EXIF metadata header
			newBytes, err = image.EncodeImage(img, format)
			if err != nil {
				l.Msg(common.ERR_IMG_ENCODE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
				return
			}

		case "gif":
			format = "webp"

			newBytes, err = image.ConvertGifToWebp(&post.Data)
			if err != nil {
				l.Msg(common.ERR_IMG_GIF_TO_WEBP_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
				return
			}

		default:
			l.Msg(common.ERR_IMG_UNKNOWN_TYPE).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
			return
		}

		// prepare the novel image's filename
		content := key + "." + format

		// upload the novel image to local storage
		//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
		if err := os.WriteFile("/opt/pix/"+content, *newBytes, 0600); err != nil {
			l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// generate thumbnails --- keep aspect ratio
		thumbImg := image.ResizeImage(img, 450)

		// encode the thumbnail back to []byte
		thumbImgData, err := image.EncodeImage(&thumbImg, format)
		if err != nil {
			l.Msg(common.ERR_IMG_THUMBNAIL_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// write the thumbnail byte stream to a file
		if err := os.WriteFile("/opt/pix/thumb_"+content, *thumbImgData, 0600); err != nil {
			l.Msg(common.ERR_IMG_SAVE_FILE_FAIL).Status(http.StatusInternalServerError).Error(err).Log().Payload(nil).Write(w)
			return
		}

		// flush post's image-related fields
		post.Figure = content
		post.Data = []byte{}
	}

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
		body, _ := json.Marshal(app.Notification{
			Title: "littr mention",
			Icon:  "/web/apple-touch-icon.png",
			Body:  callerID + " mentioned you in their post",
			Path:  "/flow/post/" + post.ID,
		})

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
	l := common.NewLogger(r, pkgName)

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
	l := common.NewLogger(r, pkgName)

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
	l := common.NewLogger(r, pkgName)

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
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Router       /posts/{postID} [get]
func getSinglePost(w http.ResponseWriter, r *http.Request) {
	l := common.NewLogger(r, pkgName)

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

	pageNoString := r.Header.Get("X-Page-No")
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
	l := common.NewLogger(r, pkgName)

	callerID, ok := r.Context().Value("nickname").(string)
	if !ok {
		l.Msg(common.ERR_CALLER_FAIL).Status(http.StatusBadRequest).Log().Payload(nil).Write(w)
		return
	}

	// parse the URI's path
	// check if diff page has been requested
	hashtag := chi.URLParam(r, "hashtag")

	pageNoString := r.Header.Get("X-Page-No")
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
