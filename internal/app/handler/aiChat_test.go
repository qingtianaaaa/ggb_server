package handler

import (
	"encoding/base64"
	"fmt"
	"ggb_server/internal/utils"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	rootPath, _ := utils.FindRootPath()
	url := "http://localhost:8080/static/upload/1752091241079782000.png"
	prefix := "http://localhost:8080"
	imgUrl := strings.Replace(url, prefix, rootPath, 1)
	log.Println(imgUrl)

	//processor := workflow.NewProcess("描述这张图片的题目",nil,nil)
	//err := processor.StartProcess(imgUrl)
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

	log.Println(imgBase64)
}
