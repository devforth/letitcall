package model

import "time"

type BusyTime struct {
	ID              string    `json:"id"`
	UserEmail       string    `json:"userEmail"`
	Start           time.Time `json:"start"`
	DurationMinutes int       `json:"durationMinutes"`
	Comment         string    `json:"comment"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (b BusyTime) End() time.Time {
	return b.Start.Add(time.Duration(b.DurationMinutes) * time.Minute)
}
