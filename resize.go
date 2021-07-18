package main

import (
	"fmt"
	"github.com/justjanne/imgconv"
	"gopkg.in/gographics/imagick.v2/imagick"
	"path/filepath"
)

func ResizeImage(config *Config, imageId string) []error {
	var err error

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err = wand.ReadImage(filepath.Join(config.SourceFolder, imageId)); err != nil {
		return []error{err}
	}
	var originalImage imgconv.ImageHandle
	if originalImage, err = imgconv.NewImage(wand); err != nil {
		return []error{err}
	}

	return runMany(len(config.Sizes), func(index int) error {
		definition := config.Sizes[index]
		path := filepath.Join(config.TargetFolder, fmt.Sprintf("%s%s", imageId, definition.Suffix))
		image := originalImage.CloneImage()
		if err := image.Crop(definition.Size); err != nil {
			return err
		}
		if err := image.Resize(definition.Size); err != nil {
			return err
		}
		if err := image.Write(config.Quality, path); err != nil {
			return err
		}
		return nil
	})
}
