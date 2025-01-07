package push

type NotificationRequest struct {
	PostID string `json:"post_id" example:"123456789000"`
}

type SubscriptionUpdateRequest struct {
	Tags []string `json:"tags" example:"reply,mention"`
}
