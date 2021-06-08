package render

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
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
}

func ActivitiesRender(w io.Writer, data ActivitiesData) error {
	t := make([][]string, 0, len(data.Items))
	for _, item := range data.Items {
		t = append(t, []string{
			fmt.Sprintf("[%s](%s)", item.Link, item.URL), item.BaseRef, item.State, item.Username, fmt.Sprintf("%s (+%d, -%d)", item.ChangeSize, item.Additions, item.Deletions), strconv.FormatInt(int64(item.Commits), 10), strconv.FormatInt(int64(item.ChangedFiles), 10),
		})
	}
	table := tablewriter.NewWriter(w)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Link", "Branch", "State", "Username", "Change Size", "Commits", "Change File"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(t)
	table.Render()
	return nil
}
