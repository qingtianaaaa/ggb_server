package schema

import (
	"encoding/json"
	"fmt"
	"log"
)

type StepFunChatCompletionRequest struct {
	Model    string      `json:"model"`
	Messages ChatMessage `json:"messages"`
	Stream   bool        `json:"stream"`
}

type ChatMessage []ContentI

type ContentI interface {
	GetContentRole() string
}

type SystemMessage struct {
	Role    string `json:"role"` //总为system
	Content string `json:"content"`
}

func (m SystemMessage) GetContentRole() string {
	if m.Role != "system" {
		log.Printf("system message role set error, want: system  actual: %s\n", m.Role)
		m.Role = "system"
	}
	return m.Role
}

type UserMessage struct {
	Role    string             `json:"role"` //总为user
	Content UserMessageContent `json:"content"`
}

func (m UserMessage) GetContentRole() string {
	if m.Role != "user" {
		log.Printf("user message role set error, want: user  actual: %s\n", m.Role)
		m.Role = "user"
	}
	return m.Role
}

type UserMessageContent []TypeI

func (ca *UserMessageContent) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*ca = make([]TypeI, len(raw))
	for i, item := range raw {
		var tmp struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(item, &tmp); err != nil {
			return err
		}

		switch tmp.Type {
		case "text":
			var t TextType
			if err := json.Unmarshal(item, &t); err != nil {
				return err
			}
			(*ca)[i] = t
		case "image_url":
			var img ImageType
			if err := json.Unmarshal(item, &img); err != nil {
				return err
			}
			(*ca)[i] = img
		case "video_url":
			var vid VideoType
			if err := json.Unmarshal(item, &vid); err != nil {
				return err
			}
			(*ca)[i] = vid
		case "input_audio":
			var aud AudioType
			if err := json.Unmarshal(item, &aud); err != nil {
				return err
			}
			(*ca)[i] = aud
		default:
			return fmt.Errorf("unknown content type: %s", tmp.Type)
		}
	}
	return nil
}

type TypeI interface {
	GetContentType() string
}

type TextType struct {
	Type string `json:"type"` //总为text
	Text string `json:"text"`
}

func (t TextType) GetContentType() string {
	return t.Type
}

type ImageType struct {
	Type     string `json:"type"` //总为image_url
	ImageUrl struct {
		Url    string `json:"url"`
		Detail string `json:"detail"`
	} `json:"image_url"`
}

func (t ImageType) GetContentType() string {
	if t.Type != "image_url" {
		log.Printf("image type set error, want: image_url  actual: %s\n", t.Type)
		t.Type = "image_url"
	}
	return t.Type
}

type VideoType struct {
	Type     string `json:"type"` //总为video_url
	VideoUrl struct {
		Url string `json:"url"`
	} `json:"video_url"`
}

func (t VideoType) GetContentType() string {
	if t.Type != "video_url" {
		log.Printf("video type set error, want: video_url  actual: %s\n", t.Type)
		t.Type = "video_url"
	}
	return t.Type
}

type AudioType struct {
	Type       string `json:"type"` //总为input_audio
	InputAudio struct {
		Data string `json:"data"`
	} `json:"input_audio"`
}

func (t AudioType) GetContentType() string {
	if t.Type != "input_audio" {
		log.Printf("audio type set error, want: input_audio  actual: %s\n", t.Type)
		t.Type = "input_audio"
	}
	return t.Type
}

type StepFunChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type StepFunChatCompletionStreamResponse struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []StepFunStreamChoice `json:"choices"`
}

type StepFunContent struct {
	Type     string
	Text     string
	ImageUrl struct {
		Url    string
		Detail string
	}
}

type StepFunStreamChoice struct {
	Delta struct {
		Content   string `json:"content"`
		Reasoning string `json:"reasoning"`
		Role      string `json:"role"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
	Index        int    `json:"index"`
}
