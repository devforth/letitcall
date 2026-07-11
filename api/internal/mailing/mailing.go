package mailing

import (
	"context"
	"time"
)

type Booking struct {
	EventName        string
	AttendeeName     string
	AttendeeEmail    string
	AttendeeTimezone string
	Notes            string
	Start            time.Time
	End              time.Time
	RecipientEmail   string
	RecipientName    string
	Timezone         string
	ManageURL        string
}

type Cancellation struct {
	Booking
	CanceledBy string
	Reason     string
}

type Sender interface {
	SendBooking(context.Context, Booking) error
	SendCancellation(context.Context, Cancellation) error
}

type Disabled struct{}

func (Disabled) SendBooking(context.Context, Booking) error {
	return nil
}

func (Disabled) SendCancellation(context.Context, Cancellation) error {
	return nil
}
