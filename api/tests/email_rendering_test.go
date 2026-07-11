package tests

import (
	"strings"
	"testing"

	"github.com/letitcall/letitcall/api/internal/mailing"
)

func TestEmailRendererEscapesHTMLAndKeepsPlainText(t *testing.T) {
	renderer, err := mailing.NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	message, err := renderer.RenderNewEvent(mailing.TemplateData{
		BrandName:        "DevForth",
		Subject:          "New Event",
		RecipientName:    "Host",
		EventName:        "Planning",
		AttendeeName:     `<script>alert("x")</script>`,
		AttendeeEmail:    "guest@example.com",
		AttendeeTimezone: "UTC",
		EventDateTime:    "12:00–12:30, Tuesday, 14 July 2026 (UTC)",
		Notes:            `<img src=x onerror=alert("x")>`,
		ManageURL:        "https://calls.example.com/event/secret",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(message.HTML, "<script>") || strings.Contains(message.HTML, "<img src=x") {
		t.Fatal("HTML email rendered active attendee content")
	}
	if !strings.Contains(message.HTML, "&lt;script&gt;") || !strings.Contains(message.HTML, "&lt;img") {
		t.Fatal("HTML email did not escape attendee content")
	}
	if !strings.Contains(message.Text, `<script>alert("x")</script>`) {
		t.Fatal("plain-text email did not retain attendee text")
	}
	if !strings.Contains(message.HTML, "DevForth") {
		t.Fatal("HTML email did not use the configured brand name")
	}
}
