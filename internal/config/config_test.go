package config

import (
	"ggb_server/internal/utils"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	path, _ := os.Getwd()
	log.Println("path: ", path)
	rootPath, _ := utils.FindRootPath()
	loadConfig(rootPath)
	log.Printf("%+v", Cfg.AIModel.DeepSeek.ApiKey)
}

func TestFuncDeep(t *testing.T) {
	// 获取调用栈信息
	pc, file, _, ok := runtime.Caller(2)
	if !ok {
		log.Fatalln("runtime.Caller() failed")
	}

	// 获取函数名
	fn := runtime.FuncForPC(pc).Name()

	log.Println("func: ", fn)
	// 获取项目根目录（假设main.go在根目录）
	_, rootFile, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(rootFile)

	log.Println("rootDir: ", rootDir)
	// 计算相对路径
	relPath, err := filepath.Rel(rootDir, filepath.Dir(file))
	if err != nil {
		log.Fatalln(relPath, err)
	}

	// 计算深度（统计路径分隔符）
	depth := 0
	if relPath != "." {
		depth = len(strings.Split(relPath, string(filepath.Separator)))
	}

	log.Println(depth)
}
