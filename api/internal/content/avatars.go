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
	avatarDataPrefix = "data:image/jpeg;base64,"
	avatarSize       = 512
	avatarTokenBytes = 4
)

type Avatar struct {
	Filename string
	contents []byte
}

type Avatars struct {
	directory string
}

func NewAvatars(dataPath string) (*Avatars, error) {
	directory := filepath.Join(dataPath, "content", "avatars")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return nil, fmt.Errorf("create avatars directory: %w", err)
	}
	return &Avatars{directory: directory}, nil
}

func (a *Avatars) Prepare(email, dataURL string) (Avatar, error) {
	if !strings.HasPrefix(dataURL, avatarDataPrefix) {
		return Avatar{}, errors.New("avatar must be a JPEG image")
	}
	contents, err := base64.StdEncoding.Strict().DecodeString(strings.TrimPrefix(dataURL, avatarDataPrefix))
	if err != nil {
		return Avatar{}, errors.New("avatar must be a valid JPEG image")
	}
	config, err := jpeg.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		return Avatar{}, errors.New("avatar must be a valid JPEG image")
	}
	if config.Width != avatarSize || config.Height != avatarSize {
		return Avatar{}, fmt.Errorf("avatar must be %d by %d pixels", avatarSize, avatarSize)
	}
	if _, err := jpeg.Decode(bytes.NewReader(contents)); err != nil {
		return Avatar{}, errors.New("avatar must be a valid JPEG image")
	}
	return prepareAvatar(email, contents)
}

func (a *Avatars) PrepareImage(email string, source image.Image) (Avatar, error) {
	bounds := source.Bounds()
	side := min(bounds.Dx(), bounds.Dy())
	if side == 0 {
		return Avatar{}, errors.New("avatar image must not be empty")
	}
	crop := image.Rect(
		bounds.Min.X+(bounds.Dx()-side)/2,
		bounds.Min.Y+(bounds.Dy()-side)/2,
		bounds.Min.X+(bounds.Dx()+side)/2,
		bounds.Min.Y+(bounds.Dy()+side)/2,
	)
	resized := image.NewRGBA(image.Rect(0, 0, avatarSize, avatarSize))
	draw.Draw(resized, resized.Bounds(), image.White, image.Point{}, draw.Src)
	xdraw.CatmullRom.Scale(resized, resized.Bounds(), source, crop, draw.Over, nil)
	var contents bytes.Buffer
	if err := jpeg.Encode(&contents, resized, &jpeg.Options{Quality: 90}); err != nil {
		return Avatar{}, fmt.Errorf("encode avatar: %w", err)
	}
	return prepareAvatar(email, contents.Bytes())
}

func (a *Avatars) Write(avatar Avatar) error {
	temporary, err := os.CreateTemp(a.directory, ".avatar-*")
	if err != nil {
		return fmt.Errorf("create temporary avatar: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(avatar.contents); err != nil {
		temporary.Close()
		return fmt.Errorf("write temporary avatar: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary avatar: %w", err)
	}
	if err := os.Rename(temporaryPath, filepath.Join(a.directory, avatar.Filename)); err != nil {
		return fmt.Errorf("store avatar: %w", err)
	}
	return nil
}

func (a *Avatars) Open(filename string) (*os.File, error) {
	if !validAvatarFilename(filename) {
		return nil, fs.ErrNotExist
	}
	return os.Open(filepath.Join(a.directory, filename))
}

func (a *Avatars) Remove(filename string) error {
	err := os.Remove(filepath.Join(a.directory, filename))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("remove avatar: %w", err)
	}
	return nil
}

func prepareAvatar(email string, contents []byte) (Avatar, error) {
	token := make([]byte, avatarTokenBytes)
	if _, err := rand.Read(token); err != nil {
		return Avatar{}, fmt.Errorf("generate avatar token: %w", err)
	}
	filename := slugEmail(email) + "-" + hex.EncodeToString(token) + ".jpg"
	return Avatar{
		Filename: filename,
		contents: contents,
	}, nil
}

func slugEmail(email string) string {
	const hexadecimal = "0123456789abcdef"
	var slug strings.Builder
	for _, character := range []byte(strings.ToLower(email)) {
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

func validAvatarFilename(filename string) bool {
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
