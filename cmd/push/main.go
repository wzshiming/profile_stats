package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/profile_stats/stats"
	"github.com/wzshiming/putingh"
)

func main() {
	ctx := context.Background()
	username := os.Getenv("GH_ID")
	token := os.Getenv("GH_TOKEN")

	owner := os.Getenv("GIT_OWNER")
	repo := os.Getenv("GIT_REPO")
	branch := os.Getenv("GIT_BRANCH")

	if username == "" {
		log.Fatal("GH_ID can not be empty")
	}
	if token == "" {
		log.Fatal("GH_TOKEN can not be empty")
	}
	if owner == "" {
		log.Fatal("GIT_OWNER can not be empty")
	}
	if repo == "" {
		log.Fatal("GIT_REPO can not be empty")
	}
	if branch == "" {
		branch = "gh-pages"
	}
	buf := bytes.NewBuffer(nil)
	src := source.NewSource(token)
	s := stats.NewStats(src)
	err := s.Get(ctx, buf, username)
	if err != nil {
		log.Fatal(err)
	}

	filename := fmt.Sprintf("%s-stats.svg", username)

	putCli := putingh.NewPutInGH(token, putingh.Config{})
	u, err := putCli.PutInGit(ctx, owner, repo, branch, filename, buf)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(u)
}
