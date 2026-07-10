package mailing

import (
	"context"
	"time"
)

type Booking struct {
	EventName      string
	AttendeeEmail  string
	Start          time.Time
	End            time.Time
	RecipientEmail string
	Timezone       string
}

type Sender interface {
	SendBooking(context.Context, Booking) error
}

type Disabled struct{}

func (Disabled) SendBooking(context.Context, Booking) error {
	return nil
}
