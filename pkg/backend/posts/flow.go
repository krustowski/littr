package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	sse "github.com/alexandrevicenzi/go-sse"
	chi "github.com/go-chi/chi/v5"
	app "github.com/maxence-charriere/go-app/v9/pkg/app"

	"go.savla.dev/littr/models"
)

// getPosts fetches posts, page spicified by a header
//
// @Summary      Get posts
// @Description  get posts
// @Tags         flow
// @Accept       json
// @Produce      json
// @Success      200  {object}  Response
// @Failure      400  {object}  Response
// @Router       /flow/ [get]
func getPosts(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")
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

	opts := pageOptions{
		CallerID: callerID,
		PageNo:   pageNo,
		FlowList: nil,
	}

	// fetch page according to the logged user
	pExport, uExport := getOnePage(opts)
	if pExport == nil || uExport == nil {
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, dumping posts"
	resp.Code = http.StatusOK

	resp.Posts = pExport
	resp.Users = uExport

	// pageSize is a constant -> see backend/pagination.go
	resp.Count = pageSize

	// write Response and logs
	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

// addNewPost adds new post
//
//	@Summary      Add new post
//	@Description  add new post
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      201  {object}  Response
//	@Failure      400  {object}  Response
//	@Failure      500  {object}  Response
//	@Router       /flow/ [post]
func addNewPost(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")
	caller, _ := r.Context().Value("nickname").(string)

	// post a new post
	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
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

		// upload to local storage
		//if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
		if err := os.WriteFile("/opt/pix/"+content, post.Data, 0600); err != nil {
			resp.Message = "backend error: couldn't save a figure to a file: " + err.Error()
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		// generate thumbanils
		if err := genThumbnails("/opt/pix/"+content, "/opt/pix/thumb_"+content); err != nil {
			resp.Message = "backend error: cannot generate the image thumbnail"
			resp.Code = http.StatusInternalServerError

			l.Println(resp.Message, resp.Code)
			resp.Write(w)
			return
		}

		post.Figure = content
		post.Data = []byte{}
	}

	if saved := setOne(FlowCache, key, post); !saved {
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
		devs, _ := getOne(SubscriptionCache, receiverName, []models.Device{})
		_, found := getOne(UserCache, receiverName, models.User{})
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

		sendNotificationToDevices(devs, body, l)
	}

	// broadcast a new post to live subscribers
	streamer.SendMessage("/api/flow/live", sse.SimpleMessage(post.Nickname))

	posts := make(map[string]models.Post)
	posts[key] = post

	resp.Message = "ok, adding new post"
	resp.Code = http.StatusCreated
	resp.Posts = posts

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// updatePostStarCount increases the star count for the given post
//
//	@Summary      Update post's star count
//	@Description  update the star count
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  Response
//	@Failure      400  {object}  Response
//	@Failure      403  {object}  Response
//	@Failure      500  {object}  Response
//	@Router       /flow/star [put]
func updatePostStarCount(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")
	callerID, _ := r.Context().Value("nickname").(string)

	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	var found bool

	if post, found = getOne(FlowCache, key, models.Post{}); !found {
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

	if saved := setOne(FlowCache, key, post); !saved {
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

// updatePost updates the specified post
//
//	@Summary      Update specified post
//	@Description  update specified post
//	@Deprecated
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  Response
//	@Failure      400  {object}  Response
//	@Failure      500  {object}  Response
//	@Router       /flow/ [put]
func updatePost(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")

	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if _, found := getOne(FlowCache, key, models.Post{}); !found {
		resp.Message = "unknown post update requested"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if saved := setOne(FlowCache, key, post); !saved {
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

// deletePost removes specified post
//
//	@Summary      Delete specified post
//	@Description  delete specified post
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  Response
//	@Failure      400  {object}  Response
//	@Failure      500  {object}  Response
//	@Router       /flow/ [delete]
func deletePost(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")

	// remove a post
	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	if err := json.Unmarshal(reqBody, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	//key := strconv.FormatInt(post.Timestamp.UnixNano(), 10)
	key := post.ID

	if deleted := deleteOne(FlowCache, key); !deleted {
		resp.Message = "cannot delete the post"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, post removed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// getUserPosts fetches posts only from specified user
//
//	@Summary      Get user posts
//	@Description  get user posts
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  Response
//	@Failure      400  {object}  Response
//	@Router       /flow/user/{nickname} [get]
func getUserPosts(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")
	callerID, _ := r.Context().Value("nickname").(string)

	// parse the URI's path
	// check if diff page has been requested
	nick := chi.URLParam(r, "nick")

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

	// mock the flowlist (nasty hack)
	flowList := make(map[string]bool)
	flowList[nick] = true

	opts := pageOptions{
		UserFlow:     true,
		UserFlowNick: nick,
		CallerID:     callerID,
		PageNo:       pageNo,
		FlowList:     flowList,
	}

	// fetch page according to the logged user
	pExport, uExport := getOnePage(opts)
	if pExport == nil || uExport == nil {
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Users = uExport
	resp.Posts = pExport

	resp.Message = "ok, dumping user's flow posts"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

// getSinglePost fetch specified post and its interaction
//
//	@Summary      Get single post
//	@Description  get single post
//	@Tags         flow
//	@Accept       json
//	@Produce      json
//	@Success      200  {object}  Response
//	@Failure      400  {object}  Response
//	@Router       /flow/post/{postNo} [get]
func getSinglePost(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	l := NewLogger(r, "flow")
	callerID, _ := r.Context().Value("nickname").(string)
	//user, _ := getOne(UserCache, callerID, models.User{})

	// parse the URI's path
	// check if diff page has been requested
	postID := chi.URLParam(r, "postNo")

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

	opts := pageOptions{
		SinglePost:   true,
		SinglePostID: postID,
		CallerID:     callerID,
		PageNo:       pageNo,
		//FlowList:   flowList,
	}

	// fetch page according to the logged user
	pExport, uExport := getOnePage(opts)
	if pExport == nil || uExport == nil {
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusBadRequest

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Users = uExport
	resp.Posts = pExport

	resp.Message = "ok, dumping single post and its interactions"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}
