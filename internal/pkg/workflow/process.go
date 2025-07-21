package workflow

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"ggb_server/internal/app/model"
	"ggb_server/internal/config"
	"ggb_server/internal/consts"
	"ggb_server/internal/pkg/workflow/aiModule"
	"ggb_server/internal/repository"
	"ggb_server/internal/utils"
	arkModel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Process struct {
	UserMessage string
	ImgUrl      string
	Flusher     http.Flusher
	W           io.Writer
	Ctx         context.Context

	config consts.Config
}

func NewProcess(userMessage string, imgUrl string, flusher http.Flusher, w io.Writer, ctx context.Context) *Process {
	return &Process{
		UserMessage: userMessage,
		ImgUrl:      imgUrl,
		Flusher:     flusher,
		W:           w,
		Ctx:         ctx,
	}
}

func (p *Process) StartProcess(db *gorm.DB, message *model.Message) error {
	rawRes, processedRes, err := p.Classify()
	workflow1 := model.Workflow{
		SessionID: message.SessionID,
		MessageID: message.ID,
		Type:      1,
		Input:     message.Message,
		Output:    rawRes,
	}
	if err != nil {
		return err
	}
	err = repository.NewWorkflowRepository().Create(db, &workflow1)
	if err != nil {
		return err
	}
	elements, err := p.ExtractElementsStream(processedRes)
	if err != nil {
		return err
	}
	input, _ := json.Marshal(processedRes)
	workflow2 := model.Workflow{
		SessionID: message.SessionID,
		MessageID: message.ID,
		RootID:    workflow1.ID,
		ParentID:  workflow1.ID,
		Type:      2,
		Input:     string(input),
		Output:    elements,
	}
	err = repository.NewWorkflowRepository().Create(db, &workflow2)
	if err != nil {
		return err
	}
	commands, err := p.GenerateGGB(elements)
	if err != nil {
		return err
	}
	workflow4ParentId := workflow2.ID
	if commands != "" {
		workflow3 := model.Workflow{
			SessionID: message.SessionID,
			MessageID: message.ID,
			RootID:    workflow1.ID,
			ParentID:  workflow2.ID,
			Type:      3,
			Input:     elements,
			Output:    commands,
		}
		err = repository.NewWorkflowRepository().Create(db, &workflow3)
		if err != nil {
			return err
		}
		workflow4ParentId = workflow3.ID
	} else {
		commands = elements
	}
	res, err := p.GenerateHTML(commands)
	if err != nil {
		return err
	}
	workflow4 := model.Workflow{
		SessionID: message.SessionID,
		MessageID: message.ID,
		RootID:    workflow1.ID,
		ParentID:  workflow4ParentId,
		Type:      4,
		Input:     elements,
		Output:    res,
	}
	err = repository.NewWorkflowRepository().Create(db, &workflow4)
	return err
}

func (p *Process) Classify() (string, map[string]string, error) {
	return p.doClassification()
}

func (p *Process) ExtractElementsStream(classifyRes map[string]string) (string, error) {
	p.config = p.lookUpClassification(classifyRes)
	if p.config.Extract.Skip {
		return "", nil
	}
	return p.doExtract(classifyRes)
}

func (p *Process) GenerateGGB(elements string) (string, error) {
	if p.config.GenGGB.Skip {
		return "", nil
	}
	return p.doGenGGB(elements)
}

func (p *Process) GenerateHTML(command string) (string, error) {
	filter := filterCommands(command)
	if p.config.GenHTML.Skip {
		return filter, nil
	}
	res, err := p.doGenHTML(filter)
	if err != nil {
		return "", err
	}
	return filterHTML(res), nil
}

func (p *Process) doClassification() (string, map[string]string, error) {
	imgUrl := p.ImgUrl
	if imgUrl != "" && (strings.Contains(imgUrl, "localhost") || strings.Contains(imgUrl, "127.0.0.1")) {
		imgPath := utils.ProcessUrl(imgUrl, config.Cfg.Static.Path)
		file, err := os.Open(imgPath)
		if err != nil {
			log.Println("Error opening file error: ", err)
			return "", nil, err
		}
		defer file.Close()
		fileContent, err := io.ReadAll(file)
		if err != nil {
			log.Println("Error reading file error: ", err)
			return "", nil, err
		}
		imgBase64 := fmt.Sprintf("data:image/%s;base64,%s", strings.ToLower(filepath.Ext(imgPath)), base64.StdEncoding.EncodeToString(fileContent))
		imgUrl = imgBase64
	}

	mapping := map[string]string{
		"model":                        string(consts.StepFuncChat1oTurbo),
		"message":                      p.UserMessage,
		strings.ToLower("imgUrl"):      imgUrl,
		strings.ToLower("processStep"): string(consts.Classify),
		strings.ToLower("contentType"): string(aiModule.Classify),
	}

	client := aiModule.NewChatCompletionClient[*aiModule.StepFunChatCompletion](mapping, p.Flusher, p.W)
	res, err := client.ChatCompletion()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", nil, err
	}
	log.Println("\n[classify content]: ", res.Content)

	classifyRes, err := p.seekForClassificationRes(res)
	if err != nil {
		return "", nil, err
	}
	return res.Content, classifyRes, nil
}

func (p *Process) seekForClassificationRes(content aiModule.Content) (map[string]string, error) {
	re := regexp.MustCompile(`(?s)\{\s*"题目"\s*:\s*"[^"]*"\s*,\s*"类型"\s*:\s*"[^"]*"\s*\}`)
	matches := re.FindString(content.Content)
	if len(matches) == 0 {
		log.Println("classification result: ", matches)
		return nil, errors.New("classification result error")
	}
	log.Println("match result: ", matches)

	res := map[string]string{}
	err := json.Unmarshal([]byte(matches), &res)
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
	log.Printf("类型： %s , 题目： %s\n\n", problemType, problem)
	return res, nil
}

func (p *Process) lookUpClassification(classify map[string]string) consts.Config {
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

func (p *Process) doExtract(classify map[string]string) (string, error) {
	problem := classify["题目"]

	mapping := map[string]string{
		"model":                        string(consts.DouBaoSeed1V6),
		"message":                      problem,
		strings.ToLower("processStep"): string(p.config.Extract.ProcessStep),
		strings.ToLower("contentType"): string(aiModule.Reasoning),
	}
	if classify["类型"] == string(consts.G3D) {
		if p.ImgUrl != "" {
			imgUrl := p.ImgUrl
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

		mapping["model"] = string(consts.StepFuncChat1oTurbo)
		mapping[strings.ToLower("thinkingType")] = string(arkModel.ThinkingTypeEnabled)
		client := aiModule.NewChatCompletionClient[*aiModule.StepFunChatCompletion](mapping, p.Flusher, p.W)

		res, err := client.ChatCompletionStream()
		if err != nil {
			log.Printf("type: %v, step: %v, content: %v\n\n", res.Type, res.Step, res.Content)
			return "", err
		}
		return res.Content, err
	}

	mapping[strings.ToLower("thinkingType")] = string(arkModel.ThinkingTypeEnabled)
	var client aiModule.ChatCompletionInterface
	client = aiModule.NewChatCompletionClient[*aiModule.DouBaoChatCompletion](mapping, p.Flusher, p.W)

	res, err := client.ChatCompletionStream()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", err
	}
	log.Println("\n[extract content]: ", res.Content)

	return res.Content, nil
}

func (p *Process) doGenGGB(elements string) (string, error) {
	filtered := filterElements(elements)
	reader := strings.Builder{}
	reader.WriteString("数学元素: \n")
	reader.WriteString(filtered)
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
	log.Println("\n[gen ggb content]: ", res.Content)
	return res.Content, nil
}

func (p *Process) doGenHTML(command string) (string, error) {
	mapping := map[string]string{
		"model":                        string(consts.DeepSeekChat),
		"message":                      command,
		strings.ToLower("processStep"): string(p.config.GenHTML.ProcessStep),
	}
	client := aiModule.NewChatCompletionClient[*aiModule.ChatCompletionClient](mapping, p.Flusher, p.W)
	res, err := client.ChatCompletion()
	if err != nil {
		log.Printf("type: %v, step: %v, content: %v\n", res.Type, res.Step, res.Content)
		return "", err
	}
	log.Println("\n[gen html content]: ", res.Content)
	return res.Content, nil
}

func filterElements(elements string) string {
	startTag := "<element_contents>"
	endTag := "</element_contents>"

	startIdx := strings.Index(elements, startTag)
	if startIdx == -1 {
		return elements
	}

	endIdx := strings.Index(elements[startIdx:], endTag)
	if endIdx == -1 {
		return elements
	}

	contentStart := startIdx + len(startTag)
	contentEnd := startIdx + endIdx
	return elements[contentStart:contentEnd]
}

func filterCommands(elements string) string {
	startTag := "<ggb_commands>"
	endTag := "</ggb_commands>"

	startIdx := strings.Index(elements, startTag)
	if startIdx == -1 {
		return elements
	}

	endIdx := strings.Index(elements[startIdx:], endTag)
	if endIdx == -1 {
		return elements
	}

	contentStart := startIdx + len(startTag)
	contentEnd := startIdx + endIdx
	return elements[contentStart:contentEnd]
}

func filterHTML(html string) string {
	startTag := "<!DOCTYPE html>"
	endTag := "</html>"

	startIdx := strings.Index(html, startTag)
	if startIdx == -1 {
		return html
	}

	endIdx := strings.Index(html[startIdx:], endTag)
	if endIdx == -1 {
		return html
	}

	contentStart := startIdx + len(startTag)
	contentEnd := startIdx + endIdx
	return html[contentStart:contentEnd]
}
