package service

import "time"

func validateTimeRange(start, end time.Time) bool {
	return start.Before(end)
}
