package workflow

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/consts"
	"ggb_server/internal/pkg/workflow/aiModule"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Process struct {
	UserMessage string
	Flusher     http.Flusher
	W           io.Writer

	config consts.Config
}

func NewProcess(userMessage string, flusher http.Flusher, w io.Writer) Process {
	return Process{
		UserMessage: userMessage,
		Flusher:     flusher,
		W:           w,
	}
}

func (p Process) StartProcess(imgUrl string) error {
	res, err := p.Classify(imgUrl)
	if err != nil {
		return err
	}
	elements, err := p.ExtractElementsStream(res, imgUrl)
	if err != nil {
		return err
	}
	commands, err := p.GenerateGGB(elements)
	if err != nil {
		return err
	}
	html, err := p.GenerateHTML(commands)
	if err != nil {
		return err
	}
	log.Println(html)
	return nil
}

func (p Process) Classify(imgUrl string) (map[string]string, error) {
	return p.doClassification(imgUrl)
}

func (p Process) ExtractElementsStream(classifyRes map[string]string, imgUrl string) (string, error) {
	p.config = p.lookUpClassification(classifyRes)
	if p.config.Extract.Skip {
		return "", nil
	}
	return p.doExtract(classifyRes, imgUrl)
}

func (p Process) GenerateGGB(elements string) (string, error) {
	if p.config.GenGGB.Skip {
		return elements, nil
	}
	return p.doGenGGB(elements)
}

func (p Process) GenerateHTML(command string) (string, error) {
	if p.config.GenHTML.Skip {
		return command, nil
	}
	return p.doGenHTML(command)
}

func (p Process) doClassification(imgUrl string) (map[string]string, error) {
	userContent := []schema.Content{
		schema.Content{
			Type: "text",
			Text: p.UserMessage,
		},
	}

	if imgUrl != "" {
		file, err := os.Open(imgUrl)
		if err != nil {
			log.Println("Error opening file error: ", err)
			return nil, err
		}
		defer file.Close()
		fileContent, err := io.ReadAll(file)
		if err != nil {
			log.Println("Error reading file error: ", err)
			return nil, err
		}
		imgBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(fileContent))

		userContent = append(userContent, schema.Content{
			Type:     "image_url",
			ImageUrl: imgBase64,
		})
	}

	userContentBytes, err := json.Marshal(userContent)
	if err != nil {
		log.Println("Error marshalling user content error: ", err)
		return nil, err
	}

	mapping := map[string]string{
		"model":                        string(consts.DeepSeekReasoner),
		"message":                      string(userContentBytes),
		strings.ToLower("processStep"): string(consts.Classify),
		strings.ToLower("contentType"): string(aiModule.Classify),
	}

	client := aiModule.NewChatCompletionClient[*aiModule.ChatCompletionClient](mapping, p.Flusher, p.W)
	res, err := client.ChatCompletionStream()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return nil, err
	}

	classifyRes, err := p.seekForClassificationRes(res)
	if err != nil {
		return nil, err
	}
	return classifyRes, nil
}

func (p Process) seekForClassificationRes(content aiModule.Content) (map[string]string, error) {
	jsonRegex := regexp.MustCompile(`(?s)(\{[^]*?\})|(\[[^]*?\])`)
	matches := jsonRegex.FindAllString(content.Content, -1) // -1 表示查找所有匹配项

	if len(matches) == 0 {
		log.Println("classification result: ", matches)
		return nil, errors.New("classification result error")
	}

	res := map[string]string{}
	err := json.Unmarshal([]byte(content.Content), &res)
	if err != nil {
		log.Println("Error unmarshalling classification result error: ", err)
		return nil, err
	}

	problemType, ok := res["类型"]
	if !ok {
		log.Println("classification failed")
		return nil, errors.New("classification failed")
	}
	problem, ok := res["题目"]
	if !ok {
		log.Println("classification failed")
		return nil, errors.New("classification failed")
	}
	log.Printf("类型： %s , 题目： %s", problemType, problem)
	return res, nil
}

func (p Process) lookUpClassification(classify map[string]string) consts.Config {
	switch classify["类型"] {
	case string(consts.G2D):
		return consts.ConfigMapping[consts.G2D]
	case string(consts.G3D):
		return consts.ConfigMapping[consts.G3D]
	case string(consts.Func):
		return consts.ConfigMapping[consts.Func]
	case string(consts.Knowledge):
		return consts.ConfigMapping[consts.Knowledge]
	case string(consts.Other):
		return consts.ConfigMapping[consts.Other]
	default:
		return consts.ConfigMapping[consts.UnknownType]
	}
}

func (p Process) doExtract(classify map[string]string, imgUrl string) (string, error) {
	problem := classify["题目"]

	mapping := map[string]string{
		"model":                        string(consts.DeepSeekReasoner),
		"message":                      problem,
		strings.ToLower("processStep"): string(p.config.Extract.ProcessStep),
		strings.ToLower("contentType"): string(aiModule.Reasoning),
	}

	var client aiModule.ChatCompletionInterface
	client = aiModule.NewChatCompletionClient[*aiModule.ChatCompletionClient](mapping, p.Flusher, p.W)

	if classify["类型"] == string(consts.G3D) {
		if imgUrl != "" {
			file, err := os.Open(imgUrl)
			if err != nil {
				log.Println("Error opening file error: ", err)
				return "", err
			}
			defer file.Close()
			fileContent, err := io.ReadAll(file)
			if err != nil {
				log.Println("Error reading file error: ", err)
				return "", err
			}
			imgBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(fileContent))
			mapping["imgUrl"] = imgBase64
		}

		mapping["model"] = string(consts.StepFuncReasoner)
		client = aiModule.NewChatCompletionClient[*aiModule.StepFunChatCompletion](mapping, p.Flusher, p.W)

		res, err := client.ChatCompletion()
		if err != nil {
			log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
			return "", err
		}
		return res.Content, err
	}

	res, err := client.ChatCompletionStream()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", err
	}

	return res.Content, nil
}

func (p Process) doGenGGB(elements string) (string, error) {
	reader := strings.Builder{}
	reader.WriteString("数学元素: \n")
	reader.WriteString(elements)
	mapping := map[string]string{
		"model":                        string(consts.DeepSeekChat),
		"message":                      reader.String(),
		strings.ToLower("processStep"): string(p.config.GenGGB.ProcessStep),
	}
	client := aiModule.NewChatCompletionClient[*aiModule.ChatCompletionClient](mapping, p.Flusher, p.W)
	res, err := client.ChatCompletion()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", err
	}
	return res.Content, nil
}

func (p Process) doGenHTML(command string) (string, error) {
	mapping := map[string]string{
		"model":                        string(consts.DeepSeekChat),
		"message":                      command,
		strings.ToLower("processStep"): string(p.config.GenHTML.ProcessStep),
	}
	client := aiModule.NewChatCompletionClient[*aiModule.ChatCompletionClient](mapping, p.Flusher, p.W)
	res, err := client.ChatCompletionStream()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", err
	}
	return res.Content, nil
}
