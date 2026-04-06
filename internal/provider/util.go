package provider

import (
	"fmt"
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

func parseDuration(s string) int {
	var hours, minutes int

	fmt.Sscanf(s, "%dh %dm", &hours, &minutes)

	return hours*60 + minutes
}
