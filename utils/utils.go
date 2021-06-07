package utils

import (
	"fmt"
	"strings"
	"time"
)

func Match(format, value string) bool {
	for _, f := range strings.Split(format, ",") {
		if match(f, value) {
			return true
		}
	}
	return false
}

func match(format string, value string) bool {
	const star = "*"
	if format == "" || format == star {
		return true
	}
	anyPrefix := strings.HasPrefix(format, star)
	anySuffix := strings.HasSuffix(format, star)
	if anyPrefix {
		format = format[1:]
	}
	if anySuffix {
		format = format[:len(format)-1]
	}
	switch {
	case anyPrefix && anySuffix:
		return strings.Contains(value, format)
	case anyPrefix && !anySuffix:
		return strings.HasSuffix(value, format)
	case !anyPrefix && anySuffix:
		return strings.HasPrefix(value, format)
	default:
		return format == value
	}
}

func ParseTimeSpan(span string, now time.Time) (time.Time, error) {
	y, m, d, err := parseTimeSpan(span)
	if err != nil {
		return time.Time{}, err
	}
	last := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).
		AddDate(-y, -m, -d)
	return last, nil
}

func parseTimeSpan(span string) (y, m, d int, err error) {
	v := 0
	u := ""

	_, err = fmt.Sscanf(span, "%d%s", &v, &u)
	if err != nil {
		return 0, 0, 0, err
	}
	u = strings.ToUpper(u)
	switch u {
	case "DAY", "DAYS", "":
		return 0, 0, v, nil
	case "MONTH", "MONTHS":
		return 0, v, 0, nil
	case "YEAR", "YEARS":
		return v, 0, 0, nil
	}
	return 0, 0, 0, fmt.Errorf("parse failure %q", span)
}

// KeyAttribute handle string, it like key:attr1=xxx:attr2=yyy
func KeyAttribute(keys []string) ([]string, map[string]map[string]string) {
	mk := map[string]map[string]string{}
	nk := make([]string, 0, len(keys))
	for _, key := range keys {
		attrs := strings.Split(key, ":")
		ma := map[string]string{}
		for _, attr := range attrs[1:] {
			kv := strings.SplitN(attr, "=", 2)
			k := kv[0]
			var v string
			if len(kv) > 1 {
				v = kv[1]
			}
			ma[k] = v
		}
		mk[attrs[0]] = ma
		nk = append(nk, attrs[0])
	}
	return nk, mk
}

func ParseTime(str string, loc *time.Location) (time.Time, error) {
	const (
		RFC3339   = time.RFC3339
		Time      = "2006-01-02T15:04:05"
		DateMonth = "2006-01"
		DateDay   = "2006-01-02"
	)

	switch len(str) {
	case len(RFC3339):
		return time.Parse(RFC3339, str)
	case len(Time):
		return time.ParseInLocation(Time, str, loc)
	case len(DateMonth):
		return time.ParseInLocation(DateMonth, str, loc)
	case len(DateDay):
		return time.ParseInLocation(DateDay, str, loc)
	}
	return time.Time{}, fmt.Errorf("can't support %q", str)
}
