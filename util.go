package main

import (
	"gopkg.in/gographics/imagick.v2/imagick"
	"math"
	"fmt"
	"errors"
)

func resize(source string, size Size, quality Quality, target string) error {
	var err error

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err = mw.ReadImage(source)
	if err != nil {
		return err
	}

	width := mw.GetImageWidth()
	height := mw.GetImageHeight()

	aspectRatio := float64(width) / float64(height)

	var nWidth, nHeight uint
	if size.Height != 0 && size.Width != 0 {
		if size.Format == SIZE_FORMAT_CONTAIN {
			if aspectRatio > 1 {
				nWidth = uint(math.Min(float64(size.Width), float64(width)))
				nHeight = uint(math.Round(float64(nWidth) / aspectRatio))
			} else {
				nHeight = uint(math.Min(float64(size.Height), float64(height)))
				nWidth = uint(math.Round(aspectRatio * float64(nHeight)))
			}
		} else if size.Format == SIZE_FORMAT_COVER {
			var targetAspectRatio = float64(size.Width) / float64(size.Height)
			var cWidth, cHeight uint
			if targetAspectRatio > aspectRatio {
				cWidth = width
				cHeight = uint(math.Round(float64(width) / targetAspectRatio))
			} else {
				cHeight = height
				cWidth = uint(math.Round(targetAspectRatio * float64(height)))
			}

			dx := int((width - cWidth) / 2)
			dy := int((height - cHeight) / 2)

			err = mw.CropImage(cWidth, cHeight, dx, dy)
			if err != nil {
				return err
			}

			nHeight = uint(math.Min(float64(size.Height), float64(cHeight)))
			nWidth = uint(math.Min(float64(size.Width), float64(cWidth)))
		} else {
			return errors.New(fmt.Sprintf("Format type not recognized: %s", size.Format))
		}
	} else if size.Height != 0 {
		nHeight = uint(math.Min(float64(size.Height), float64(height)))
		nWidth = uint(math.Round(aspectRatio * float64(nHeight)))
	} else if size.Width != 0 {
		nWidth = uint(math.Min(float64(size.Width), float64(width)))
		nHeight = uint(math.Round(float64(nWidth) / aspectRatio))
	} else {
		nWidth = width
		nHeight = height
	}

	if (width != nWidth) || (height != nHeight) {
		err = mw.TransformImageColorspace(imagick.COLORSPACE_RGB)
		if err != nil {
			return err
		}

		err = mw.ResizeImage(nWidth, nHeight, imagick.FILTER_LANCZOS, 1)
		if err != nil {
			return err
		}

		err = mw.TransformImageColorspace(imagick.COLORSPACE_SRGB)
		if err != nil {
			return err
		}
	}

	if quality.CompressionQuality != 0 {
		mw.SetImageCompressionQuality(quality.CompressionQuality)
	}

	if len(quality.SamplingFactors) != 0 {
		mw.SetSamplingFactors(quality.SamplingFactors)
	}

	err = mw.WriteImage(target)
	return err
}