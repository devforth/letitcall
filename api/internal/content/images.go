package content

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
)

const (
	squareImageSize = 512
	imageTokenBytes = 4
)

type Image struct {
	Filename string
	contents []byte
}

type squareImages struct {
	directory string
	noun      string
}

func newSquareImages(dataPath, folder, noun string) (*squareImages, error) {
	directory := filepath.Join(dataPath, "content", folder)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create %s directory: %w", folder, err)
	}
	return &squareImages{directory: directory, noun: noun}, nil
}

func (i *squareImages) Prepare(subject, dataURL string) (Image, error) {
	prefix := "data:image/jpeg;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		return Image{}, fmt.Errorf("%s must be a JPEG image", i.noun)
	}
	contents, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(dataURL, prefix))
	if err != nil {
		return Image{}, fmt.Errorf("%s must be a valid JPEG image", i.noun)
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		return Image{}, fmt.Errorf("%s must be a valid JPEG image", i.noun)
	}
	if config.Width != squareImageSize || config.Height != squareImageSize {
		return Image{}, fmt.Errorf("%s must be %d by %d pixels", i.noun, squareImageSize, squareImageSize)
	}
	if _, err := jpeg.Decode(bytes.NewReader(contents)); err != nil {
		return Image{}, fmt.Errorf("%s must be a valid JPEG image", i.noun)
	}
	return prepareImage(subject, contents)
}

func (i *squareImages) PrepareImage(subject string, source image.Image) (Image, error) {
	bounds := source.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	if side == 0 {
		return Image{}, fmt.Errorf("%s image must not be empty", i.noun)
	}
	crop := image.Rect(
		bounds.Min.X+(bounds.Dx()-side)/2,
		bounds.Min.Y+(bounds.Dy()-side)/2,
		bounds.Min.X+(bounds.Dx()+side)/2,
		bounds.Min.Y+(bounds.Dy()+side)/2,
	)
	resized := image.NewRGBA(image.Rect(0, 0, squareImageSize, squareImageSize))
	draw.Draw(resized, resized.Bounds(), image.White, image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), source, crop, draw.Over, nil)
	var contents bytes.Buffer
	if err := jpeg.Encode(&contents, resized, &jpeg.Options{Quality: 90}); err != nil {
		return Image{}, fmt.Errorf("encode %s: %w", i.noun, err)
	}
	return prepareImage(subject, contents.Bytes())
}

func (i *squareImages) Write(image Image) error {
	temporary, err := os.CreateTemp(i.directory, "."+i.noun+"-*")
	if err != nil {
		return fmt.Errorf("create temporary %s: %w", i.noun, err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(image.contents); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary %s: %w", i.noun, err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary %s: %w", i.noun, err)
	}
	if err := os.Rename(temporaryPath, filepath.Join(i.directory, image.Filename)); err != nil {
		return fmt.Errorf("store %s: %w", i.noun, err)
	}
	return nil
}

func (i *squareImages) Open(filename string) (*os.File, error) {
	if !validImageFilename(filename) {
		return nil, fs.ErrNotExist
	}
	return os.Open(filepath.Join(i.directory, filename))
}

func (i *squareImages) Remove(filename string) error {
	err := os.Remove(filepath.Join(i.directory, filename))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("remove %s: %w", i.noun, err)
	}
	return nil
}

func prepareImage(subject string, contents []byte) (Image, error) {
	token := make([]byte, imageTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return Image{}, fmt.Errorf("generate image token: %w", err)
	}
	return Image{
		Filename: slugImageSubject(subject) + "-" + hex.EncodeToString(token) + ".jpg",
		contents: contents,
	}, nil
}

func slugImageSubject(subject string) string {
	const hexadecimal = "0123456789abcdef"
	var slug strings.Builder
	for _, character := range []byte(strings.ToLower(subject)) {
		switch {
		case character == '@':
			slug.WriteString("__")
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9', character == '.', character == '-', character == '+':
			slug.WriteByte(character)
		default:
			slug.WriteByte('~')
			slug.WriteByte(hexadecimal[character>>4])
			slug.WriteByte(hexadecimal[character&0x0f])
		}
	}
	return slug.String()
}

func validImageFilename(filename string) bool {
	if len(filename) <= len(".jpg") || !strings.HasSuffix(filename, ".jpg") {
		return false
	}
	for _, character := range strings.TrimSuffix(filename, ".jpg") {
		if !((character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || character == '.' || character == '-' || character == '_' || character == '+' || character == '~') {
			return false
		}
	}
	return true
}
