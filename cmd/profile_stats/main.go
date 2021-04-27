package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/wzshiming/profile_stats/activities"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/profile_stats/stats"
	"github.com/wzshiming/putingh"
	"github.com/wzshiming/xmlinjector"
)

var key = []byte("PROFILE_STATS")

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
	for _, uri := range uris {
		buf.Reset()
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

		origin := buf.Bytes()
		data, err := xmlinjector.Inject(key, origin, func(args, origin []byte) []byte {
			tag := reflect.StructTag(args)
			template, ok := tag.Lookup("template")
			if !ok || template == "" {
				return errInfo("no template")
			}

			switch template {
			default:
				return errInfo(fmt.Sprintf("not support template %q", template))
			case "updatedat":
				return []byte(time.Now().Format(time.RFC3339))
			case "activities":
				username, ok := tag.Lookup("username")
				if !ok || username == "" {
					return errInfo("no username")
				}
				activity := activities.NewActivities(src)
				buf := bytes.NewBuffer([]byte("\n"))
				err = activity.Get(ctx, buf, username)
				if err != nil {
					return errInfo(err.Error())
				}
				buf.WriteByte('\n')
				return buf.Bytes()
			case "stats":
				username, ok := tag.Lookup("username")
				if !ok || username == "" {
					return errInfo("no username")
				}
				stat := stats.NewStats(src)
				buf := bytes.NewBuffer([]byte("\n"))
				err = stat.Get(ctx, buf, username)
				if err != nil {
					return errInfo(err.Error())
				}
				buf.WriteByte('\n')
				return buf.Bytes()
			}
		})

		if err != nil {
			log.Println(err, uri)
			continue
		}

		if bytes.Equal(origin, data) {
			log.Println("no need to update", uri)
			continue
		}

		out, err := putCli.PutIn(ctx, uri, bytes.NewBuffer(data))
		if err != nil {
			log.Println(err, uri)
			continue
		}
		log.Printf("updated %s: %s", uri, out)
	}
	return
}

func errInfo(msg string) []byte {
	return []byte(fmt.Sprintf("\n<!-- error: %s ->\n", msg))
}
