package main

import (
	"fmt"
	"github.com/justjanne/imgconv"
	"gopkg.in/gographics/imagick.v3/imagick"
	"path/filepath"
	"time"
)

func ResizeImage(config *Config, imageId string) []error {
	var err error

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	startRead := time.Now().UTC()
	if err = wand.ReadImage(filepath.Join(config.SourceFolder, imageId)); err != nil {
		return []error{err}
	}
	var originalImage imgconv.ImageHandle
	if originalImage, err = imgconv.NewImage(wand); err != nil {
		return []error{err}
	}
	trackTimeSince(imageProcessDurationRead, startRead)

	return runMany(len(config.Sizes), func(index int) error {
		definition := config.Sizes[index]
		path := filepath.Join(config.TargetFolder, fmt.Sprintf("%s%s", imageId, definition.Suffix))
		startClone := time.Now().UTC()
		image := originalImage.CloneImage()
		startCrop := trackTimeSince(imageProcessDurationClone, startClone)
		if err := image.Crop(definition.Size); err != nil {
			return err
		}
		startResize := trackTimeSince(imageProcessDurationCrop, startCrop)
		if err := image.Resize(definition.Size); err != nil {
			return err
		}
		startWrite := trackTimeSince(imageProcessDurationResize, startResize)
		if err := image.Write(config.Quality, path); err != nil {
			return err
		}
		trackTimeSince(imageProcessDurationWrite, startWrite)
		return nil
	})
}
