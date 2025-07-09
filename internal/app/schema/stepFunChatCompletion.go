package schema

type StepFunChatCompletionRequest struct {
	Model    string
	Messages []Message
	Stream   bool
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
