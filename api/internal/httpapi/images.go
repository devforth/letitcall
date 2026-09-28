package httpapi

import (
	"github.com/letitcall/letitcall/api/internal/content"
	"github.com/letitcall/letitcall/api/internal/model"
)

type imageUploadRequest struct {
	Rendered string            `json:"rendered"`
	Original string            `json:"original"`
	Editor   model.ImageEditor `json:"editor"`
}

type preparedImage struct {
	rendered content.Image
	original content.Image
	source   *model.ImageSource
}

func prepareImageUpload(
	request imageUploadRequest,
	previous *model.ImageSource,
	prepareRendered func(string) (content.Image, error),
	prepareOriginal func(string) (content.Image, error),
) (preparedImage, error) {
	rendered, err := prepareRendered(request.Rendered)
	if err != nil {
		return preparedImage{}, err
	}
	path := ""
	if previous != nil {
		path = previous.Path
	}
	var original content.Image
	if request.Original != "" {
		original, err = prepareOriginal(request.Original)
		if err != nil {
			return preparedImage{}, err
		}
		path = original.Filename
	}
	return preparedImage{
		rendered: rendered,
		original: original,
		source:   &model.ImageSource{Path: path, Editor: request.Editor},
	}, nil
}

type imageFiles interface {
	Write(content.Image) error
	Remove(string) error
}

func writePreparedImage(files imageFiles, image preparedImage) error {
	if err := files.Write(image.rendered); err != nil {
		return err
	}
	if image.original.Filename == "" {
		return nil
	}
	if err := files.Write(image.original); err != nil {
		_ = files.Remove(image.rendered.Filename)
		return err
	}
	return nil
}

func removePreparedImage(files imageFiles, image preparedImage) {
	if image.rendered.Filename != "" {
		_ = files.Remove(image.rendered.Filename)
	}
	if image.original.Filename != "" {
		_ = files.Remove(image.original.Filename)
	}
}
