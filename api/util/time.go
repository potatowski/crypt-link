package util

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func ParseTime(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func FormatTime(t time.Time, layout string) string {
	return t.UTC().Format(layout)
}

func IsExpired(t time.Time) bool {
	return NowUTC().After(t)
}
