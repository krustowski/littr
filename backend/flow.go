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

func GetPosts() *[]Post {
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

func AddPost(content string) bool {
	var posts *[]Post = GetPosts()
	var post Post = Post{
		Nickname:  "random",
		Content:   content,
		Timestamp: time.Now(),
	}

	*posts = append(*posts, post)

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
