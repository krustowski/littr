package backend

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"time"
)

type Posts struct {
	Posts []Post `json:"posts"`
}

func getPosts() *[]Post {
	var posts Posts

	dat, err := os.ReadFile("/opt/data/posts.json")
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	err = json.Unmarshal(dat, &posts)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	// order posts by timestamp DESC
	sort.SliceStable(posts.Posts, func(i, j int) bool {
		return posts.Posts[i].Timestamp.After(posts.Posts[j].Timestamp)
	})

	return &posts.Posts
}

func addPost(post Post) bool {
	var posts *[]Post = getPosts()

	var newPost Post = Post{
		Nickname:  post.Nickname,
		Content:   post.Content,
		Timestamp: time.Now(),
	}

	*posts = append(*posts, newPost)

	postsToWrite := &Posts{
		Posts: *posts,
	}

	jsonData, err := json.Marshal(postsToWrite)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	err = os.WriteFile("/opt/data/posts.json", jsonData, 0644)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	return true
}
