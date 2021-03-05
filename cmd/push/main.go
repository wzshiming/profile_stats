package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/profile_stats/stats"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_ID")
	if username == "" {
		log.Fatal("GITHUB_ID can not be empty")
	}
	if token == "" {
		log.Fatal("GITHUB_TOKEN can not be empty")
	}
	buf := bytes.NewBuffer(nil)
	src := source.NewSource(token)
	s := stats.NewStats(src)
	err := s.Get(ctx, buf, username)
	if err != nil {
		log.Fatal(err)
	}
	repo := "profile_stats"
	filename := fmt.Sprintf("%s-stats.svg", username)
	u, err := src.UploadGist(ctx, username, repo, filename, buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(u)
}
