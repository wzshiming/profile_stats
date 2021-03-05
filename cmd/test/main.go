package main

import (
	"context"
	"log"
	"os"

	"github.com/wzshiming/profile_stats/stats"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("GITHUB_TOKEN")
	username := os.Getenv("GITHUB_ID")
	s := stats.NewStats(token)
	err := s.Get(ctx, os.Stdout, username)
	if err != nil {
		log.Fatal(err)
	}
}
