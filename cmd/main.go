package main

import (
	"ggb_server/internal/app"
	"ggb_server/internal/config"
)

func init() {
	config.LoadConfig()
}

func main() {
	app.Start()
}
