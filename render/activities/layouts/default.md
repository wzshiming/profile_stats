| Link | Ref | State | Change Size | Commits | Change File |
| - | - | - | - | - | - |
{{range $i, $item := .Items}}
| [{{with $item.Link}}{{.}}{{else}}UNKNOWN{{end}}]({{with $item.URL}}{{.}}{{end}}) | {{.BaseRef}} | {{.State}} | {{.ChangeSize}} (<font color="#56d364">+{{.Additions}}</font>, <font color="#f85149">-{{.Deletions}}</font>) | {{.Commits}} | {{.ChangedFiles}} |
{{end}}
