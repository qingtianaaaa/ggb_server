package schema

type ChatRequest struct {
	UserId       string `json:"userId"`
	Message      string `json:"message" binding:"required"`
	SessionId    int    `json:"sessionId"`
	ParentId     int    `json:"parentId"`
	Title        string `json:"title"`
	MessageOrder int    `json:"messageOrder"`
	ImageUrl     string `json:"imageUrl"`
}
