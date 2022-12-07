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
	var post Post = Post{
		Nickname:  "random",
		Content:   content,
		Timestamp: time.Now(),
	}

	d := NewData()
	d.Read("posts").SetData("posts", post).Write("posts")

	return true
}
