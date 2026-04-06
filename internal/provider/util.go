package provider

import (
	"os"
	"time"
)

func loadMock(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func parseWithTimezone(dt string, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, err
	}

	return time.ParseInLocation("2006-01-02T15:04:05", dt, loc)
}
