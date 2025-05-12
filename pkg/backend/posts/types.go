package posts

type PostCreateRequest struct {
	Type       string `json:"type" example:"post" enums:"post,poll,img"`
	ReplyToID  string `json:"reply_to_id" example:"1234567890000"`
	Content    string `json:"content" example:"a very random post's content"`
	FigureName string `json:"figure_name" example:"example.jpg"`
	FigureData []byte `json:"figure_data" swaggertype:"string" format:"base64" example:"base64 encoded data"`
}

type PostPagingRequest struct {
	PageNo       int
	PagingSize   int
	HideReplies  bool
	SinglePost   bool
	SinglePostID string
	Hashtag      string
	SingleUser   bool
	SingleUserID string
}
