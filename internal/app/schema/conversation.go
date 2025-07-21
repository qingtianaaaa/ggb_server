package schema

type CreateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

type CreateConversationResponse struct {
	ID    uint   `json:"id"`
	Title string `json:"title"`
}
