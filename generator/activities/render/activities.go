package render

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/wzshiming/profile_stats/source"
)

type ActivitiesData struct {
	Items []ActivitiesItem
}

type ActivitiesItem struct {
	URL          string
	Username     string
	Link         string
	Title        string
	BaseRef      string
	State        string
	Additions    int
	Deletions    int
	Commits      int
	ChangedFiles int
	ChangeSize   string
	CreatedAt    time.Time
	ClosedAt     time.Time
	MergedAt     time.Time
	UpdatedAt    time.Time
	Labels       []string
}

func ActivitiesRender(w io.Writer, data ActivitiesData) error {
	t := make([][]string, 0, len(data.Items))
	for _, item := range data.Items {
		label := strings.Join(item.Labels, "<br/>")
		label = strings.ReplaceAll(label, " ", "&nbsp;")
		state := item.State
		switch state {
		case string(source.PullRequestStateMerged):
			mergedAt := formatTime(item.MergedAt)
			state = fmt.Sprintf("Merged<br/>%s", mergedAt)
		case string(source.PullRequestStateOpen):
			createdAt := formatTime(item.CreatedAt)
			updatedAt := formatTime(item.UpdatedAt)
			if createdAt == updatedAt {
				state = fmt.Sprintf("Open<br/>%s", createdAt)
			} else {
				state = fmt.Sprintf("Open<br/>%s<br/>%s", createdAt, updatedAt)
			}
		case string(source.PullRequestStateClosed):
			closedAt := formatTime(item.ClosedAt)
			state = fmt.Sprintf("Closed<br/>%s", closedAt)
		}

		change := fmt.Sprintf("%s<br/>(+%d,-%d)/%d/%d", item.ChangeSize, item.Additions, item.Deletions, item.Commits, item.ChangedFiles)
		link := fmt.Sprintf("[%s](%s)", item.Link, item.URL)
		t = append(t, []string{
			link, item.BaseRef, state, item.Username, change, label,
		})
	}
	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Link", "Branch", "State", "Username", "Change Size/Commits/File", "Labels"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(t)
	table.Render()
	return nil
}

func formatTime(t time.Time) string {
	return t.Local().Format("2006-01-02")
}
