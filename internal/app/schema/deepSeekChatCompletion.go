package schema

type Message struct {
	Content string `json:"content"` //可为结构体也可为string 如果为StepFunContent结构体列表 提前序列化
	Role    string `json:"role"`
	Name    string `json:"name"`
}

type Content struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	ImageUrl string `json:"image_url"`
}

type DeepSeekRequest struct {
	Messages         []Message       `json:"messages" required:"true"`
	Model            string          `json:"model" required:"true"`
	FrequencyPenalty int             `json:"frequency_penalty"`
	MaxTokens        int             `json:"max_tokens"` //[1,2^32-1]
	PresencePenalty  int             `json:"presence_penalty"`
	ResponseFormat   *ResponseFormat `json:"response_format,omitempty"` // omitempty 当值为零值时不序列化
	Stop             []string        `json:"stop,omitempty"`            //最多16个string
	Stream           bool            `json:"stream"`
	StreamOptions    *StreamOptions  `json:"stream_options,omitempty"`
	Temperature      float64         `json:"temperature"`           // 0～2 and default 1
	TopP             float64         `json:"top_p"`                 // (0,1]
	Tools            []interface{}   `json:"tools,omitempty"`       // 比较复杂 见官网解释
	ToolChoice       string          `json:"tool_choice,omitempty"` // 比较复杂 见官网解释
	Logprobs         bool            `json:"logprobs"`
	TopLogProbs      *int            `json:"top_logprobs,omitempty"` // 0～20的整数
}

type ResponseFormat struct {
	Type interface{} `json:"type"` //string or json object
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type DeepSeekResponse struct { //非流式输出
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"` //unix时间戳 单位秒
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	Usage             Usage    `json:"usage"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type Choice struct {
	Index        int         `json:"index"`
	Message      RespMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
	Logprobs     *Logprobs   `json:"logprobs,omitempty"`
}

type RespMessage struct {
	Content          string   `json:"content"`
	ReasoningContent string   `json:"reasoning_content"`
	Role             string   `json:"role"`
	ToolCalls        ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Usage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	PromptCacheHitTokens    int `json:"prompt_cache_hit_tokens"`
	PromptCacheMissTokens   int `json:"prompt_cache_miss_tokens"`
	CompletionTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

type Logprobs struct {
	Content []struct {
		Token       string  `json:"token"`
		Logprob     float64 `json:"logprob"`
		Bytes       []int   `json:"bytes"`
		TopLogprobs []struct {
			Token   string  `json:"token"`
			Logprob float64 `json:"logprob"`
			Bytes   []int   `json:"bytes"`
		} `json:"top_logprobs"`
	} `json:"content"`
}

type DeepSeekStreamResponse struct {
	ID                string         `json:"id"`
	Choices           []StreamChoice `json:"choices"`
	Created           int64          `json:"created"`
	Model             string         `json:"model"`
	SystemFingerprint string         `json:"system_fingerprint"`
	Object            string         `json:"object"`
}

type StreamChoice struct {
	Delta struct {
		Content          string `json:"content"`
		ReasoningContent string `json:"reasoning_content"`
		Role             string `json:"role"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
	Index        int    `json:"index"`
}
