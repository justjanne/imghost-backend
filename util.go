package main

import (
	"errors"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"math"
)

func resize(wand *imagick.MagickWand, wandLinear *imagick.MagickWand, size Size, quality Quality, target string) error {
	var err error
	var mw *imagick.MagickWand

	colorSpace := mw.GetImageColorspace()

	if size.Width == 0 && size.Height == 0 {
		mw = wand.Clone()
		defer mw.Destroy()
	} else {
		mw = wandLinear.Clone()
		defer mw.Destroy()

		width := mw.GetImageWidth()
		height := mw.GetImageHeight()

		aspectRatio := float64(width) / float64(height)

		var nWidth, nHeight uint
		if size.Height != 0 && size.Width != 0 {
			if size.Format == sizeFormatContain {
				if aspectRatio > 1 {
					nWidth = uint(math.Min(float64(size.Width), float64(width)))
					nHeight = uint(math.Round(float64(nWidth) / aspectRatio))
				} else {
					nHeight = uint(math.Min(float64(size.Height), float64(height)))
					nWidth = uint(math.Round(aspectRatio * float64(nHeight)))
				}
			} else if size.Format == sizeFormatCover {
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
			err = mw.ResizeImage(nWidth, nHeight, imagick.FILTER_LANCZOS, 1)
			if err != nil {
				return err
			}

			err = mw.TransformImageColorspace(colorSpace)
			if err != nil {
				return err
			}
		}
	}

	if quality.CompressionQuality != 0 {
		mw.SetImageCompressionQuality(quality.CompressionQuality)
	}

	if len(quality.SamplingFactors) != 0 {
		mw.SetSamplingFactors(quality.SamplingFactors)
	}

	mw.StripImage()
	mw.SetImageColorspace(colorSpace)

	err = mw.WriteImage(target)

	return err
}
