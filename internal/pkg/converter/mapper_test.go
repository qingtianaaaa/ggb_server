package converter

import (
	"fmt"
	"testing"
)

type ChatRequest struct {
	UserId       string `json:"userId"`
	Message      string `json:"message" binding:"required"`
	SessionId    uint   `json:"sessionId"`
	ParentId     uint   `json:"parentId"`
	Title        string `json:"title"`
	MessageOrder uint   `json:"messageOrder"`
	ImageUrl     string `json:"imageUrl"`
}

type Chat struct {
	UserId    string `json:"userId"`
	Message   string `json:"message" binding:"required"`
	SessionId uint   `json:"sessionId"`
	ParentId  uint   `json:"parentId"`
	Title     string `json:"title" mapper:"omitempty"`
	ImageUrl  string `json:"imageUrl" mapper:"-"`
	Resource  string `json:"resource"`
}

type ChatResponse struct {
	Message string `json:"message"`
	Title   string `json:"title"`
}

func Test_01(t *testing.T) {
	req := ChatRequest{
		UserId:       "userId",
		Message:      "message",
		SessionId:    1,
		ParentId:     1,
		Title:        "title",
		MessageOrder: 1,
		ImageUrl:     "imageUrl",
	}

	var chat Chat
	err := Convert(req, &chat)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", chat)
}

func Test_02(t *testing.T) {
	type Chat2 struct {
		UserID  string `json:"userId" mapper:"Message"`
		Message string `json:"message" mapper:"UserId"`
		Session uint   `json:"sessionId" mapper:"SessionId"`
		Parent  uint   `json:"parentId"`
		Title   string `json:"title"`
	}
	req := ChatRequest{
		UserId:       "userId",
		Message:      "message",
		SessionId:    2,
		ParentId:     1,
		Title:        "title",
		MessageOrder: 1,
		ImageUrl:     "imageUrl",
	}
	var chat Chat2
	err := Convert(req, &chat)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v\n", chat)
}
