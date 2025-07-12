package aiModule

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/consts"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
)

type ChatCompletionClient struct {
	Ctx          context.Context
	Model        consts.DeepSeekModel
	ProcessStep  consts.ProcessStep
	Message      string
	Flusher      http.Flusher
	StreamWriter io.Writer
	ContentType  Type
}

func NewChatCompletionClient[T ChatCompletionInterface](mapping map[string]string, flusher http.Flusher, w io.Writer) T {
	var instance T

	elemType := reflect.TypeOf(instance).Elem()
	clientValue := reflect.New(elemType)

	if flusher != nil {
		if fField := clientValue.Elem().FieldByName("Flusher"); fField.IsValid() && fField.CanSet() {
			fField.Set(reflect.ValueOf(flusher))
		}
	}
	if w != nil {
		if swField := clientValue.Elem().FieldByName("StreamWriter"); swField.IsValid() && swField.CanSet() {
			swField.Set(reflect.ValueOf(w))
		}
	}

	if mapping == nil {
		return clientValue.Interface().(T)
	}

	structType := clientValue.Elem().Type()
	for i := 0; i < structType.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := clientValue.Elem().Field(i)

		if fieldType.Type.Kind() == reflect.String && fieldValue.CanSet() {
			key := strings.ToLower(fieldType.Name)
			if v, ok := mapping[key]; ok {
				fieldValue.SetString(v)
			}
		}
	}

	return clientValue.Interface().(T)
}

func (g ChatCompletionClient) ChatCompletion() (Content, error) {
	chatCompletionReq := schema.DeepSeekRequest{
		Model: string(g.Model),
		Messages: []schema.Message{
			schema.Message{
				Role:    "system",
				Content: consts.ProcessStepMapping[g.ProcessStep],
			},
			schema.Message{
				Role:    "user",
				Content: g.Message,
			},
		},
		TopP:      1,
		MaxTokens: 1024,
		Stream:    false, //非流式
	}
	payload, err := json.Marshal(chatCompletionReq)
	if err != nil {
		log.Println("json marshal error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, consts.DeepSeekChatCompletionUrl, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("new request error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+consts.DeepSeekApiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Print("send request error:", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("API returned error: ", string(body))
		return Content{
			Type:    Error,
			Content: string(body),
			Step:    g.ProcessStep,
		}, errors.New("API returned error: " + resp.Status)
	}

	reader, _ := io.ReadAll(resp.Body)
	response := schema.DeepSeekResponse{}
	fullResponse := strings.Builder{}
	_ = json.Unmarshal(reader, &response)
	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		fullResponse.WriteString(content)
		writeSSEEvent(g.StreamWriter, g.Flusher, content) //也以流式形式返回前端
	}
	return Content{
		Type:    OutputContent,
		Content: fullResponse.String(),
		Step:    g.ProcessStep,
	}, nil
}

func (g ChatCompletionClient) ChatCompletionStream() (Content, error) {
	chatCompletionReq := schema.DeepSeekRequest{
		Model: string(g.Model),
		Messages: []schema.Message{
			{
				Role:    "system",
				Content: consts.ProcessStepMapping[g.ProcessStep],
			},
			{
				Role:    "user",
				Content: g.Message,
			},
		},
		TopP:      1,
		MaxTokens: 1024,
		Stream:    true, // 流式
	}

	reqBody, err := json.Marshal(chatCompletionReq)
	if err != nil {
		log.Println("Error marshaling request: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}

	req, err := http.NewRequest(http.MethodPost, consts.DeepSeekChatCompletionUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println("Error creating request: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+consts.DeepSeekApiKey)
	req.Header.Set("Accept", "text/event-stream")
	log.Println("api key : ", consts.DeepSeekApiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request: ", err)
		return Content{
			Type:    Error,
			Content: err.Error(),
			Step:    g.ProcessStep,
		}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Println("API returned non-OK : ", string(bodyBytes))
		return Content{
			Type:    Error,
			Content: string(bodyBytes),
			Step:    g.ProcessStep,
		}, errors.New("API returned non-OK : " + string(bodyBytes))
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

		if strings.HasPrefix(line, "data: ") {
			jsonStr := strings.TrimPrefix(line, "data: ")
			jsonStr = strings.TrimSpace(jsonStr)

			if jsonStr == "[DONE]" {
				break
			}

			var chunk schema.DeepSeekStreamResponse
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				log.Printf("unmarshal chunk error: %v, content: %s", err, jsonStr)
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.ReasoningContent != "" {
				reasoningContent := chunk.Choices[0].Delta.ReasoningContent
				formatContent := Content{
					Type:    Reasoning,
					Step:    g.ProcessStep,
					Content: reasoningContent,
				}
				fmt.Print(reasoningContent)
				jsonBody, _ := json.Marshal(formatContent)
				writeSSEEvent(g.StreamWriter, g.Flusher, string(jsonBody))
				//fullResponse.WriteString(string(jsonBody))
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				content := chunk.Choices[0].Delta.Content
				formatContent := Content{
					Type:    OutputContent,
					Step:    g.ProcessStep,
					Content: content,
				}
				jsonBody, _ := json.Marshal(formatContent)
				writeSSEEvent(g.StreamWriter, g.Flusher, string(jsonBody))
				fullResponse.WriteString(content)
			}
		}
	}

	return Content{
		Type:    OutputContent,
		Content: fullResponse.String(),
		Step:    g.ProcessStep,
	}, nil
}

func writeSSEEvent(w io.Writer, flusher http.Flusher, data string) {
	if flusher == nil || w == nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
