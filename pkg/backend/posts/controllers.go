package posts

import (
	"encoding/json"
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
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /posts/ [get]
func getPosts(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)

	pageNo := 0

	pageNoString := r.Header.Get("X-Flow-Page-No")
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		resp.Message = "page No has to be specified as integer/number"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	pageNo = page

	if callerID == "" {
		resp.Message = "callerID header cannot be blank!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	opts := PageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,
	}

	// fetch page according to the logged user
	pExport, uExport := GetOnePage(opts)
	if pExport == nil || uExport == nil {
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, dumping posts"
	resp.Code = http.StatusOK

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	// TODO: use DTO
	for key, user := range uExport {
		user.Passphrase = ""
		user.PassphraseHex = ""
		user.Email = ""

		if user.Nickname != callerID {
			user.FlowList = nil
			user.ShadeList = nil
			user.RequestList = nil
		}

		uExport[key] = user
	}

	resp.Posts = pExport
	resp.Users = uExport

	resp.Key = callerID

	// pageSize is a constant -> see backend/pagination.go
	resp.Count = PageSize

	// write Response and logs
	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// addNewPost adds new post
//
// @Summary      Add new post
// @Description  add new post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      201  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /posts/ [post]
func addNewPost(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	caller, _ := r.Context().Value("nickname").(string)

	// post a new post
	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if post.Nickname == "" {
		post.Nickname = caller
	}

	// check the post forgery possibility
	if caller != post.Nickname {
		resp.Message = "invalid author spotted --- one can post under their authenticated name only"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	post.ID = key
	post.Timestamp = timestamp

	// uploadedFigure handling
	if post.Data != nil && post.Figure != "" {
		fileExplode := strings.Split(post.Figure, ".")
		extension := fileExplode[len(fileExplode)-1]

		content := key + "." + extension

		// decode image from []byte stream
		img, format, err := decodeImage(post.Data)
		if err != nil {
			resp.Message = "backend error: cannot decode given byte stream"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// fix the image orientation for decoded image
		img, err = fixOrientation(img, post.Data)
		if err != nil {
			resp.Message = "backend error: cannot fix image's oriantation: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// re-encode the image
		newBytes, err := encodeImage(img, format)
		if err != nil {
			resp.Message = "backend error: cannot re-encode the novel image"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// remove EXIF metadata
		_, newBytes, err = removeExif(newBytes, format)
		if err != nil {
			resp.Message = "backend error: cannot remove EXIF metadata"
			resp.Code = http.StatusBadRequest

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// upload to local storage
		//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
		if err := os.WriteFile("/opt/pix/"+content, newBytes, 0600); err != nil {
			resp.Message = "backend error: couldn't save a figure to a file: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// generate thumbanils
		thumbImg := resizeImage(img, 350)

		// encode the thumbnail back to []byte
		thumbImgData, err := encodeImage(thumbImg, format)
		if err != nil {
			resp.Message = "backend error: cannot encode thumbnail back to byte stream"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// write the thumbnail byte stream to a file
		if err := os.WriteFile("/opt/pix/thumb_"+content, thumbImgData, 0600); err != nil {
			resp.Message = "backend error: couldn't save a figure to a file: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		post.Figure = content
		post.Data = []byte{}
	}

	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		resp.Message = "backend error: cannot save new post (cache error)"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		if receiverName == caller {
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
			Body:  caller + " mentioned you in their post",
			Path:  "/flow/post/" + post.ID,
		})

		push.SendNotificationToDevices(receiverName, devs, body, l)
	}

	// broadcast a new post to live subscribers
	Streamer.SendMessage("/api/v1/posts/live", sse.SimpleMessage("post,"+post.Nickname))

	posts := make(map[string]models.Post)
	posts[key] = post

	resp.Message = "ok, adding new post"
	resp.Code = http.StatusCreated
	resp.Posts = posts

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updatePostStarCount increases the star count for the given post.
//
// @Summary      Update post's star count
// @Description  update the star count
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      403  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /posts/{postID}/star [patch]
func updatePostStarCount(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)

	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	var found bool

	if post, found = db.GetOne(db.FlowCache, key, models.Post{}); !found {
		resp.Message = "unknown post update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if post.Nickname == callerID {
		resp.Message = "one cannot rate their own post(s)"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// increment the star count by 1
	post.ReactionCount++

	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		resp.Message = "cannot update post"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, star count incremented"
	resp.Code = http.StatusOK

	resp.Posts = make(map[string]models.Post)
	resp.Posts[key] = post

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updatePost updates the specified post.
//
// @Summary      Update specified post
// @Description  update specified post
// @Deprecated
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      403  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /posts/{postID} [put]
func updatePost(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)

	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if _, found := db.GetOne(db.FlowCache, key, models.Post{}); !found {
		resp.Message = "unknown post update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if post.Nickname != callerID {
		resp.Message = "one cannot update foreign post(s)"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if saved := db.SetOne(db.FlowCache, key, post); !saved {
		resp.Message = "cannot update post"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, post updated"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// deletePost removes specified post.
//
// @Summary      Delete specified post
// @Description  delete specified post
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Failure      403  {object}  common.Response
// @Failure      500  {object}  common.Response
// @Router       /posts/{postID} [delete]
func deletePost(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)

	// remove a post
	var post models.Post

	if err := common.UnmarshalRequestData(r, &post); err != nil {
		resp.Message = "input read error: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if post.Nickname != callerID {
		resp.Message = "one cannot delete foreign post(s)"
		resp.Code = http.StatusForbidden

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if deleted := db.DeleteOne(db.FlowCache, key); !deleted {
		resp.Message = "cannot delete the post"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// delete associated image and thumbnail
	if post.Figure != "" {
		err := os.Remove("/opt/pix/thumb_" + post.Figure)
		if err != nil {
			resp.Message = "cannot remove thumbnail associated to the post"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		err = os.Remove("/opt/pix/" + post.Figure)
		if err != nil {
			resp.Message = "cannot remove image associated to the post"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}
	}

	resp.Message = "ok, post removed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
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
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)
	//user, _ := getOne(UserCache, callerID, models.User{})

	// parse the URI's path
	// check if diff page has been requested
	postID := chi.URLParam(r, "postID")

	pageNo := 0

	pageNoString := r.Header.Get("X-Flow-Page-No")
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		resp.Message = "page No has to be specified as integer/number"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// flush email addresses
	for key, user := range uExport {
		if key == callerID {
			continue
		}
		user.Email = ""
		uExport[key] = user
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	resp.Users = uExport
	resp.Posts = pExport

	resp.Key = callerID

	resp.Message = "ok, dumping single post and its interactions"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// fetchHashtaggedPosts
//
// @Summary      Get hashtagged post list
// @Description  get hashtagged post list
// @Tags         posts
// @Accept       json
// @Produce      json
// @Success      200  {object}  common.Response
// @Failure      400  {object}  common.Response
// @Router       /posts/hashtag/{hashtag} [get]
func fetchHashtaggedPosts(w http.ResponseWriter, r *http.Request) {
	resp := common.Response{}
	l := common.NewLogger(r, pkgName)
	callerID, _ := r.Context().Value("nickname").(string)

	// parse the URI's path
	// check if diff page has been requested
	hashtag := chi.URLParam(r, "hashtag")

	pageNoString := r.Header.Get("X-Flow-Page-No")
	page, err := strconv.Atoi(pageNoString)
	if err != nil {
		resp.Message = "page No has to be specified as integer/number"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
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
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	// flush email addresses
	for key, user := range uExport {
		if key == callerID {
			continue
		}
		user.Email = ""
		uExport[key] = user
	}

	// hack: include caller's models.User struct
	if caller, ok := db.GetOne(db.UserCache, callerID, models.User{}); !ok {
		resp.Message = "cannot fetch such callerID-named user"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	} else {
		uExport[callerID] = caller
	}

	resp.Users = uExport
	resp.Posts = pExport

	resp.Key = callerID

	resp.Message = "ok, dumping hashtagged posts and their parent posts"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}
