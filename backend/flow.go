package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.savla.dev/littr/config"
	"go.savla.dev/littr/models"
)

const (
	pagination int = 50
)

func convertMapToArray[T any](m map[string]T, reverseOutput bool) ([]string, []T) {
	var keys = []string{}
	var vals = []T{}

	for key, val := range m {
		keys = append(keys, key)
		vals = append(vals, val)
	}

	if reverseOutput {
		reverse(keys)
		//reverse(vals)
	}

	return keys, vals
}

func getPostsPaged(w http.ResponseWriter, r *http.Request) {
}

func getPosts(w http.ResponseWriter, r *http.Request) {
	resp := response{}
	l := NewLogger(r, "flow")
	noteUsersActivity(r.Header.Get("X-API-Caller-ID"))

	user, ok := getOne(UserCache, r.Header.Get("X-API-Caller-ID"), models.User{})
	if !ok {
		// cache error, or caller does not exist...
		// TODO handle response and log
		return
	}

	// fetch the flow + users and combine them into one response
	// those variables are both of type map[string]T
	allPosts, _ := getAll(FlowCache, models.Post{})
	allUsers, _ := getAll(UserCache, models.User{})

	// pagination draft
	// + only select N latest posts for such user according to their FlowList
	// + include previous posts to a reply
	// + only include users mentioned

	// prepare a reversed array of post keys
	// reverse it to get them ordered by time DESC
	//keys, _ := convertMapToArray(posts, true)
	//keys := []string{}
	posts := []models.Post{}
	num := 0

	for _, post := range allPosts {
		// check the caller's flow list, skip on unfollowed, or unknown user
		if value, found := user.FlowList[post.Nickname]; !found || !value {
			continue
		}

		num++
		//keys = append(keys, key)
		posts = append(posts, post)
	}

	// order posts by timestamp DESC
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp.After(posts[j].Timestamp)
	})

	// cut the <pagination>*2 number of keys only
	var part []models.Post

	if len(posts) > pagination*2 {
		part = posts[0:(pagination * 2)]
	} else {
		part = posts
	}

	// loop through the array and manually include other posts too
	// watch for users as well
	pExport := make(map[string]models.Post)
	uExport := make(map[string]models.User)

	num = 0
	for _, post := range part {
		// increase the count of posts
		num++

		// export one (1) post
		pExport[post.ID] = post
		uExport[post.Nickname] = allUsers[post.Nickname]

		// we can have multiple keys from a single post -> its interractions
		repKey := post.ReplyToID
		if repKey != "" {
			num++
			pExport[repKey] = allPosts[repKey]

			// export previous user too
			nick := allPosts[repKey].Nickname
			uExport[nick] = allUsers[nick]
		}

		if num > pagination {
			break
		}
	}

	resp.Message = "ok, dumping posts"
	resp.Code = http.StatusOK
	resp.Posts = pExport
	resp.Users = uExport
	resp.Count = pagination

	l.Println(resp.Message, resp.Code)
	resp.Write(w)
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

	posts, _ := getAll(FlowCache, models.Post{})

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
