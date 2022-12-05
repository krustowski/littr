package backend

import (
	"sort"
	"time"
)

var posts []Post = []Post{
	{Nickname: "system", Content: "welcome onboard bruh, lit ngl", Timestamp: 1669997122},
	{Nickname: "krusty", Content: "idk sth ig", Timestamp: 1669997543},
}

func GetPosts() *[]Post {
	// order posts by timestamp DESC
	sort.SliceStable(posts, func(i, j int) bool {
		return posts[i].Timestamp > posts[j].Timestamp
	})

	return &posts
}

func AddPost(content string) bool {
	var post Post = Post{
		Nickname:  "random",
		Content:   content,
		Timestamp: int(time.Now().Unix()),
	}

	posts = append(posts, post)

	return true
}
