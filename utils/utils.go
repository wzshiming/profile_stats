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
