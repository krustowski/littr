package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

func getPosts(w http.ResponseWriter, r *http.Request) {
	resp := response{}
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

	// write response and logs
	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

func addNewPost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")

	// post a new post
	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	timestamp := time.Now()
	key := strconv.FormatInt(timestamp.UnixNano(), 10)

	post.ID = key
	post.Timestamp = timestamp

	// uploadedFigure handling
	if post.Data != nil && post.Content != "" {
		fileExplode := strings.Split(post.Content, ".")
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
		genThumbnails("/opt/pix/"+content, "/opt/pix/thumb_"+content)

		post.Content = content
		post.Data = []byte{}
	}

	if saved := setOne(FlowCache, key, post); !saved {
		resp.Message = "backend error: cannot save new post (cache error)"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	posts := make(map[string]models.Post)
	posts[key] = post

	resp.Message = "ok, adding new post"
	resp.Code = http.StatusCreated
	resp.Posts = posts

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func updatePostStarCount(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")

	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

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

	// increment the star count by 1
	post.ReactionCount++

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

func updatePost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")

	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

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

func deletePost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")

	// remove a post
	var post models.Post

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		resp.Message = "backend error: cannot read input stream: " + err.Error()
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	data := config.Decrypt([]byte(os.Getenv("APP_PEPPER")), reqBody)

	if err := json.Unmarshal(data, &post); err != nil {
		resp.Message = "backend error: cannot unmarshall fetched data: " + err.Error()
		resp.Code = http.StatusInternalServerError

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

func getUserPosts(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")
	callerID := r.Header.Get("X-API-Caller-ID")

	// parse the URI's path
	// check if diff page has been requested
	nick := chi.URLParam(r, "nick")

	pageNo := 0

	// mock the flowlist (nasty hack)
	flowList := make(map[string]bool)
	flowList[callerID] = true

	opts := pageOptions{
		UserFlow:     true,
		UserFlowNick: nick,
		CallerID:     callerID,
		PageNo:       pageNo,
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

	resp.Message = "ok, dumping user's flow posts"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}

func getSinglePost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")
	callerID := r.Header.Get("X-API-Caller-ID")

	pageNo := 0

	// mock the flowlist (nasty hack)
	flowList := make(map[string]bool)
	flowList[callerID] = true

	opts := pageOptions{
		SinglePost: true,
		CallerID:   callerID,
		PageNo:     pageNo,
		FlowList:   flowList,
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

	resp.Message = "ok, dumping single post and its interactions"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}
