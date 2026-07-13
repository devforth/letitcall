package content

import (
	"image"
	"os"
)

type Avatar = Image

type Avatars struct {
	images *squareImages
}

func NewAvatars(dataPath string) (*Avatars, error) {
	images, err := newSquareImages(dataPath, "avatars", "avatar")
	if err != nil {
		return nil, err
	}
	return &Avatars{images: images}, nil
}

func (a *Avatars) Prepare(email, dataURL string) (Avatar, error) {
	return a.images.Prepare(email, dataURL)
}

func (a *Avatars) PrepareImage(email string, source image.Image) (Avatar, error) {
	return a.images.PrepareImage(email, source)
}

func (a *Avatars) Write(avatar Avatar) error {
	return a.images.Write(avatar)
}

func (a *Avatars) Open(filename string) (*os.File, error) {
	return a.images.Open(filename)
}

func (a *Avatars) Remove(filename string) error {
	return a.images.Remove(filename)
}
