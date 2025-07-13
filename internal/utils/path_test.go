package utils

import (
	"log"
	"path/filepath"
	"testing"
)

func TestPath(t *testing.T) {
	url := "http://localhost:8081/static/upload/1752324138383213000.jpg"
	ext := filepath.Ext(url)
	log.Println(ext)
}
