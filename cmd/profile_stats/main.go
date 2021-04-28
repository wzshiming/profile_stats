package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"

	"github.com/wzshiming/profile_stats/generator"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/putingh"
)

func main() {
	ctx := context.Background()
	token := os.Getenv("GH_TOKEN")
	uris := os.Args[1:]
	Update(ctx, token, uris...)
}

func Update(ctx context.Context, token string, uris ...string) {
	buf := bytes.NewBuffer(nil)
	putCli := putingh.NewPutInGH(token, putingh.Config{})
	src := source.NewSource(token)
	regi := generator.NewHandler(src)
	for _, uri := range uris {
		buf.Reset()
		local := !strings.Contains(uri, ":/")
		if local {
			f, err := os.Open(uri)
			if err != nil {
				log.Println(err, uri)
				continue
			}
			_, err = buf.ReadFrom(f)
			f.Close()
			if err != nil {
				log.Println(err, uri)
				continue
			}
		} else {
			r, err := putCli.GetFrom(ctx, uri)
			if err != nil {
				log.Println(err, uri)
				continue
			}
			_, err = buf.ReadFrom(r)
			if err != nil {
				log.Println(err, uri)
				continue
			}
		}

		origin := buf.Bytes()
		data, err := regi.Handle(ctx, origin)
		if err != nil {
			log.Println(err, uri)
			continue
		}

		if bytes.Equal(origin, data) {
			log.Println("no need to update", uri)
			continue
		}

		if local {
			err = os.WriteFile(uri, data, 0666)
			if err != nil {
				log.Println(err, uri)
				continue
			}
			log.Printf("updated %s", uri)
		} else {
			out, err := putCli.PutIn(ctx, uri, bytes.NewBuffer(data))
			if err != nil {
				log.Println(err, uri)
				continue
			}
			log.Printf("updated %s: %s", uri, out)
		}
	}
	return
}
