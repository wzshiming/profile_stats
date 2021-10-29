package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/wzshiming/profile_stats/generator"
	"github.com/wzshiming/profile_stats/source"
	"github.com/wzshiming/putingh"
)

const selfRepo = "https://github.com/wzshiming/profile_stats"

func main() {
	ctx := context.Background()
	token := os.Getenv("GH_TOKEN")
	warningExit, _ := strconv.ParseBool(os.Getenv("WARNING_EXIT"))
	interval, _ := time.ParseDuration(os.Getenv("INTERVAL"))
	retry, _ := strconv.ParseInt(os.Getenv("RETRY"), 0, 64)
	tmp := os.Getenv("TMP_DIR")
	uris := os.Args[1:]
	err := Update(ctx, token, tmp, interval, int(retry), warningExit, uris...)
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
}

func Update(ctx context.Context, token, tmp string, interval time.Duration, retry int, warningExit bool, uris ...string) error {
	putCli := putingh.NewPutInGH(token,
		putingh.WithGitCommitMessage(func(owner, repo, branch, name, path string) string {
			return fmt.Sprintf(`Automatic update %s

For details see %s
`, name, selfRepo)
		}),
		putingh.WithTmpDir(tmp),
	)

	buf := bytes.NewBuffer(nil)
	src := source.NewSource(token, tmp, interval, retry)
	regi := generator.NewHandler(src)
	for _, uri := range uris {
		buf.Reset()
		local := !strings.Contains(uri, ":/")
		if local {
			f, err := os.Open(uri)
			if err != nil {
				return fmt.Errorf("open %s: %w", uri, err)
			}
			_, err = buf.ReadFrom(f)
			f.Close()
			if err != nil {
				return fmt.Errorf("read %s: %w", uri, err)
			}
		} else {
			r, err := putCli.GetFrom(ctx, uri)
			if err != nil {
				return fmt.Errorf("open %s: %w", uri, err)
			}
			_, err = buf.ReadFrom(r)
			if err != nil {
				return fmt.Errorf("read %s: %w", uri, err)
			}
		}

		origin := buf.Bytes()
		data, warnings, err := regi.Handle(ctx, origin)
		if err != nil {
			return fmt.Errorf("handle %s: %w", uri, err)
		}

		if len(warnings) != 0 {
			for _, warning := range warnings {
				log.Println(warning)
			}
			if warningExit {
				return fmt.Errorf("warning exit")
			}
		}

		if bytes.Equal(origin, data) {
			log.Println("no need to update", uri)
			continue
		}

		if local {
			err = os.WriteFile(uri, data, 0666)
			if err != nil {
				return fmt.Errorf("write %s: %w", uri, err)
			}
			log.Printf("updated %s", uri)
		} else {
			out, err := putCli.PutIn(ctx, uri, bytes.NewBuffer(data))
			if err != nil {
				return fmt.Errorf("write %s: %w", uri, err)
			}
			log.Printf("updated %s: %s", uri, out)
		}
	}
	return nil
}
