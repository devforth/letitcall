package content

import "os"

type Logo = Image

type Logos struct {
	images *squareImages
}

func NewLogos(dataPath string) (*Logos, error) {
	images, err := newSquareImages(dataPath, "logos", "logo")
	if err != nil {
		return nil, err
	}
	return &Logos{images: images}, nil
}

func (l *Logos) Prepare(dataURL string) (Logo, error) {
	return l.images.Prepare("logo", dataURL)
}

func (l *Logos) Write(logo Logo) error {
	return l.images.Write(logo)
}

func (l *Logos) Open(filename string) (*os.File, error) {
	return l.images.Open(filename)
}

func (l *Logos) Remove(filename string) error {
	return l.images.Remove(filename)
}
