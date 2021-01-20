package lib

import (
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"path/filepath"
)

func goroutineWrapper(
	errorChannel chan error,
	wand *imagick.MagickWand,
	config *Config, path string,
	definition SizeDefinition,
	parameters ImageParameters,
) {
	errorChannel <- ProcessImageInstance(
		wand,
		definition.Size,
		config.Quality,
		parameters,
		path,
	)
}

func ResizeImage(config *Config, imageId string) []error {
	var err error
	errorChannel := make(chan error)

	wand := imagick.NewMagickWand()
	defer wand.Destroy()

	if err = wand.SetOption("png:include-chunk", "bKGD,cHRM,iCCP"); err != nil {
		return []error{err}
	}
	if err = wand.SetOption("png:exclude-chunk", "EXIF,iTXt,tEXt,zTXt,date"); err != nil {
		return []error{err}
	}

	if err = wand.ReadImage(filepath.Join(config.SourceFolder, imageId)); err != nil {
		return []error{err}
	}

	_ = wand.AutoOrientImage()

	colorSpace := wand.GetImageColorspace()
	if colorSpace == imagick.COLORSPACE_UNDEFINED {
		colorSpace = imagick.COLORSPACE_SRGB
	}

	profiles := map[string][]byte{}
	for _, name := range wand.GetImageProfiles("*") {
		profiles[name] = []byte(wand.GetImageProfile(name))
	}

	parameters := ImageParameters{
		depth:      wand.GetImageDepth(),
		profiles:   profiles,
		colorspace: colorSpace,
	}

	if err = wand.SetImageDepth(16); err != nil {
		panic(err)
	}
	if err = wand.ProfileImage("icc", WorkingColorspace); err != nil {
		panic(err)
	}

	for _, definition := range config.Sizes {
		path := filepath.Join(config.TargetFolder, fmt.Sprintf("%s%s", imageId, definition.Suffix))
		go goroutineWrapper(errorChannel, wand, config, path, definition, parameters)
	}

	errors := make([]error, 0)
	for i := 0; i < len(config.Sizes); i++ {
		err := <-errorChannel
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
