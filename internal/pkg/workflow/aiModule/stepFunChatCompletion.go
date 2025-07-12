package aiModule

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/consts"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type StepFunChatCompletion struct {
	Model        consts.StepFunModel
	ProcessStep  consts.ProcessStep
	Message      string
	Flusher      http.Flusher
	StreamWriter io.Writer
	ImgUrl       string
}

func (s StepFunChatCompletion) ChatCompletion() (Content, error) {
	content := []schema.StepFunContent{
		schema.StepFunContent{
			Type: "text",
			Text: s.Message,
		},
	}
	if s.ImgUrl != "" {
		content = append(content, schema.StepFunContent{
			Type: "image_url",
			ImageUrl: struct {
				Url    string
				Detail string
			}{Url: s.ImgUrl, Detail: "high"},
		})
	}

	contentByte, err := json.Marshal(content)
	if err != nil {
		log.Println("Error marshalling user content error: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}

	chatCompletionReq := schema.StepFunChatCompletionRequest{
		Model: string(s.Model),
		Messages: []schema.Message{
			schema.Message{
				Role:    "system",
				Content: consts.ProcessStepMapping[s.ProcessStep],
			},
			schema.Message{
				Role:    "user",
				Content: string(contentByte),
			},
		},
		Stream: false,
	}
	payload, err := json.Marshal(chatCompletionReq)
	if err != nil {
		log.Println("json marshal error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, consts.StepFunChatCompletionUrl, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("new request error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+consts.StepFunApiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Print("send request error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("API returned error: ", string(body))
		return Content{
			Type:    Error,
			Content: string(body),
			Step:    s.ProcessStep,
		}, errors.New("API returned error: " + resp.Status)
	}

	reader, _ := io.ReadAll(resp.Body)
	response := schema.StepFunChatCompletionResponse{}
	fullResponse := strings.Builder{}
	_ = json.Unmarshal(reader, &response)
	if len(response.Choices) > 0 {
		outputContent := response.Choices[0].Message.Content
		fullResponse.WriteString(outputContent)
		writeSSEEvent(s.StreamWriter, s.Flusher, outputContent) //也以流式形式返回前端
	}
	return Content{
		Type:    OutputContent,
		Content: fullResponse.String(),
		Step:    s.ProcessStep,
	}, nil
}

func (s StepFunChatCompletion) ChatCompletionStream() (Content, error) {
	content := []schema.StepFunContent{
		schema.StepFunContent{
			Type: "text",
			Text: s.Message,
		},
	}
	if s.ImgUrl != "" {
		content = append(content, schema.StepFunContent{
			Type: "image_url",
			ImageUrl: struct {
				Url    string
				Detail string
			}{Url: s.ImgUrl, Detail: "high"},
		})
	}

	contentByte, err := json.Marshal(content)
	if err != nil {
		log.Println("Error marshalling user content error: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}

	chatCompletionReq := schema.StepFunChatCompletionRequest{
		Model: string(s.Model),
		Messages: []schema.Message{
			schema.Message{
				Role:    "system",
				Content: consts.ProcessStepMapping[s.ProcessStep],
			},
			schema.Message{
				Role:    "user",
				Content: string(contentByte),
			},
		},
		Stream: true,
	}

	reqBody, err := json.Marshal(chatCompletionReq)
	if err != nil {
		log.Println("Error marshalling user content error: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}

	req, err := http.NewRequest(http.MethodPost, consts.StepFunChatCompletionUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    s.ProcessStep,
		}, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "text/event-stream")
	req.Header.Add("Authorization", "Bearer "+consts.StepFunApiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Content{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Content{
			Type:    Error,
			Content: "",
			Step:    s.ProcessStep,
		}, errors.New(resp.Status)
	}

	fullResponse := strings.Builder{}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println("Error reading stream: ", err)
		}

		if strings.HasPrefix(line, "data:") {
			jsonStr := strings.Trim(line, "data: ")
			jsonStr = strings.TrimSpace(jsonStr)

			if jsonStr == "[DONE]" {
				break
			}

			var chunk schema.StepFunChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Reasoning != "" {
				reasoningContent := chunk.Choices[0].Delta.Reasoning
				formatContent := Content{
					Type:    Reasoning,
					Step:    s.ProcessStep,
					Content: reasoningContent,
				}
				jsonBody, _ := json.Marshal(formatContent)
				writeSSEEvent(s.StreamWriter, s.Flusher, string(jsonBody))
				//fullResponse.WriteString(reasoningContent)
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				output := chunk.Choices[0].Delta.Content
				formatContent := Content{
					Type:    OutputContent,
					Step:    s.ProcessStep,
					Content: output,
				}
				log.Printf("format content: %+v\n", output)
				jsonBody, _ := json.Marshal(formatContent)
				writeSSEEvent(s.StreamWriter, s.Flusher, string(jsonBody))
				fullResponse.WriteString(output)
			}
		}
	}
	return Content{
		Type:    OutputContent,
		Content: fullResponse.String(),
		Step:    s.ProcessStep,
	}, nil
}
