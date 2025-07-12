package aiModule

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"ggb_server/internal/app/schema"
	"ggb_server/internal/consts"
	"ggb_server/internal/utils"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_reflect(t *testing.T) {
	client := &ChatCompletionClient{}

	typeC := reflect.TypeOf(client)
	valueC := reflect.ValueOf(client)

	structFieldType := typeC
	structFieldValue := valueC
	if reflect.TypeOf(client).Kind() == reflect.Pointer {
		structFieldType = structFieldType.Elem()
		structFieldValue = structFieldValue.Elem()
	}

	for i := 0; i < structFieldType.NumField(); i++ {
		fieldValue := structFieldValue.Field(i)
		fieldType := structFieldType.Field(i)
		log.Println("filedType.Type.Name(): ", fieldType.Type.Name())
		log.Println("filedType.Type.Kind(): ", fieldType.Type.Kind())
		log.Println("fieldType.Name: ", fieldType.Name)
		log.Println("fieldValue.CanSet(): ", fieldValue.CanSet())
		if fieldType.Type.Kind() == reflect.String && fieldValue.CanSet() {
			fieldValue.SetString("hello")
			log.Println("fieldValue.Interface(): ", fieldValue.Interface())
		}
		log.Println("----")
	}
}

func Test_Chat(t *testing.T) {
	rootPath, _ := utils.FindRootPath()
	url := "http://localhost:8080/static/upload/1752091241079782000.png"
	prefix := "http://localhost:8080"
	imgUrl := strings.Replace(url, prefix, rootPath, 1)
	file, err := os.Open(imgUrl)
	if err != nil {
		log.Println("Error opening file error: ", err)
	}
	defer file.Close()
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file error: ", err)
	}
	imgBase64 := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(fileContent))
	originalSize := len(fileContent)
	encodedSize := len(base64.StdEncoding.EncodeToString(fileContent))
	fmt.Printf("原始文件大小: %d 字节 (%.2f KB)\n", originalSize, float64(originalSize)/1024)
	fmt.Printf("Base64编码后大小: %d 字节 (%.2f KB)\n", encodedSize, float64(encodedSize)/1024)
	fmt.Printf("大小增加比例: %.2f%%\n", float64(encodedSize-originalSize)/float64(originalSize)*100)
	userContent := []schema.Content{
		schema.Content{
			Type: "text",
			Text: "描述这道题目, 同时使用中文回答 不要用英语",
		},
		schema.Content{
			Type:     "image_url",
			ImageUrl: imgBase64,
		},
	}

	userContentBytes, err := json.Marshal(userContent)
	mapping := map[string]string{
		"model":                        string(consts.StepFuncChat),
		"message":                      string(userContentBytes),
		strings.ToLower("imgUrl"):      imgBase64,
		strings.ToLower("processStep"): string(consts.Classify),
		strings.ToLower("contentType"): string(Classify),
	}
	client := NewChatCompletionClient[*StepFunChatCompletion](mapping, nil, nil)

	res, err := client.ChatCompletion()

	if err != nil {
		log.Println("err: ", err)
		return
	}
	log.Println("res: ", res.Content)
}
