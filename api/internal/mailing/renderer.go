package mailing

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"path"
	texttemplate "text/template"
)

//go:embed templates/rendered templates/text
var templateFiles embed.FS

type TemplateData struct {
	BrandName        string
	Subject          string
	RecipientName    string
	EventName        string
	AttendeeName     string
	AttendeeEmail    string
	AttendeeTimezone string
	EventDateTime    string
	Notes            string
	CanceledBy       string
	Reason           string
	ManageURL        string
}

type Message struct {
	Subject string
	HTML    string
	Text    string
}

type Renderer struct {
	html map[string]*htmltemplate.Template
	text map[string]*texttemplate.Template
}

func NewRenderer() (*Renderer, error) {
	renderer := &Renderer{html: make(map[string]*htmltemplate.Template), text: make(map[string]*texttemplate.Template)}
	for _, name := range []string{"new-event", "cancellation"} {
		htmlSource, err := templateFiles.ReadFile(path.Join("templates/rendered", name+".html"))
		if err != nil {
			return nil, fmt.Errorf("read rendered %s email template: %w", name, err)
		}
		renderer.html[name], err = htmltemplate.New(name).Parse(string(htmlSource))
		if err != nil {
			return nil, fmt.Errorf("parse rendered %s email template: %w", name, err)
		}
		textSource, err := templateFiles.ReadFile(path.Join("templates/text", name+".txt"))
		if err != nil {
			return nil, fmt.Errorf("read text %s email template: %w", name, err)
		}
		renderer.text[name], err = texttemplate.New(name).Parse(string(textSource))
		if err != nil {
			return nil, fmt.Errorf("parse text %s email template: %w", name, err)
		}
	}
	return renderer, nil
}

func (r *Renderer) RenderNewEvent(data TemplateData) (Message, error) {
	return r.render("new-event", data)
}

func (r *Renderer) RenderCancellation(data TemplateData) (Message, error) {
	return r.render("cancellation", data)
}

func (r *Renderer) render(name string, data TemplateData) (Message, error) {
	var htmlBody bytes.Buffer
	if err := r.html[name].Execute(&htmlBody, data); err != nil {
		return Message{}, fmt.Errorf("render %s HTML email: %w", name, err)
	}
	var textBody bytes.Buffer
	if err := r.text[name].Execute(&textBody, data); err != nil {
		return Message{}, fmt.Errorf("render %s text email: %w", name, err)
	}
	return Message{Subject: data.Subject, HTML: htmlBody.String(), Text: textBody.String()}, nil
}
