package config

import (
	"fmt"
	"ggb_server/internal/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Server struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type Database struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"` // 分钟
}

type Static struct {
	Path string `mapstructure:"path"`
}

type Log struct {
	Level               string `mapstructure:"level"`
	FilePath            string `mapstructure:"file_path"`
	EnableConsoleOutput bool   `mapstructure:"enable_console_output"`
}

type AIModel struct {
	DeepSeek struct {
		ApiKey string `mapstructure:"apiKey"`
	} `mapstructure:"deepseek"`
	StepFun struct {
		ApiKey string `mapstructure:"apiKey"`
	} `mapstructure:"stepfun"`
	DouBao struct {
		ApiKey string `mapstructure:"apiKey"`
	} `mapstructure:"doubao"`
}

type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"db"`
	JWT      JWT      `mapstructure:"jwt"`
	Static   Static   `mapstructure:"static"`
	Log      Log      `mapstructure:"log"`
	AIModel  AIModel  `mapstructure:"ai_model"`
}

var (
	once sync.Once
	Cfg  *Config
)

func init() {
	rootPath, _ := findRootPath()
	loadConfig(rootPath)
}

func LoadConfig() {
	log.Println("loading config file")
}

func loadConfig(dir string) {
	once.Do(func() {
		var config Config

		viper.AddConfigPath(dir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("read config file failed: %v\n", err))
		}

		if err := viper.Unmarshal(&config); err != nil {
			panic(fmt.Errorf("parse config file failed: %v\n", err))
		}
		name, _ := utils.FindProjectName()
		if config.Server.Name != name {
			panic(fmt.Errorf("service names in config.yml and go.mod are inconsistent. Please make them match"))
		}

		Cfg = &config
	})
}

func findRootPath() (string, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current working directory: %w", err)
	}
	for {
		goModPath := filepath.Join(startDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return startDir, nil
		}

		parentDir := filepath.Dir(startDir)
		if parentDir == startDir {
			return "", fmt.Errorf("go.mod not found in or above %s", startDir)
		}
		startDir = parentDir
	}
}
