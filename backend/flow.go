package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"

	"github.com/go-chi/chi/v5"
)

const (
	pagination int = 50
)

func getPosts(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	pageNo := 0

	// check if diff page has been requested
	page := chi.URLParam(r, "pageNo")
	if page != "" {
		num, err := strconv.Atoi(page)
		if err == nil {
			pageNo = num
		}
	}

	// fetch page according to the logged user
	pExport, uExport := getOnePage(pageNo, r.Header.Get("X-API-Caller-ID"))
	if pExport == nil || uExport == nil {
		resp.Message = "error while requesting more page, one exported map is nil!"
		resp.Code = http.StatusInternalServerError

		l.Println(resp.Message, resp.Code)
		resp.Write(w)
		return
	}

	resp.Message = "ok, dumping posts"
	resp.Code = http.StatusOK

	resp.Posts = pExport
	resp.Users = uExport
	resp.Count = pageSize

	// write response and logs
	l.Println(resp.Message, resp.Code)
	resp.Write(w)
	return
}

func addNewPost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

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

func updatePost(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "flow",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

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
	l := Logger{
		CallerID:  r.Header.Get("X-API-Caller-ID"),
		IPAddress: r.Header.Get("X-Real-IP"),
		//IPAddress:  r.RemoteAddr,
		Method:     r.Method,
		Route:      r.URL.String(),
		WorkerName: "flow",
		Version:    r.Header.Get("X-App-Version"),
	}
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

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

	timestamp := time.Now()
	key = strconv.FormatInt(timestamp.UnixNano(), 10)

	resp.Message = "ok, post removed"
	resp.Code = http.StatusOK

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
}
