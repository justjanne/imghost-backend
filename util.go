package main

import (
	"errors"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"math"
)

func colorSpaceName(colorSpace imagick.ColorspaceType) string {
	if colorSpace == imagick.COLORSPACE_UNDEFINED {
		return "COLORSPACE_UNDEFINED"
	} else if colorSpace == imagick.COLORSPACE_RGB {
		return "COLORSPACE_RGB"
	} else if colorSpace == imagick.COLORSPACE_GRAY {
		return "COLORSPACE_GRAY"
	} else if colorSpace == imagick.COLORSPACE_TRANSPARENT {
		return "COLORSPACE_TRANSPARENT"
	} else if colorSpace == imagick.COLORSPACE_OHTA {
		return "COLORSPACE_OHTA"
	} else if colorSpace == imagick.COLORSPACE_LAB {
		return "COLORSPACE_LAB"
	} else if colorSpace == imagick.COLORSPACE_XYZ {
		return "COLORSPACE_XYZ"
	} else if colorSpace == imagick.COLORSPACE_YCBCR {
		return "COLORSPACE_YCBCR"
	} else if colorSpace == imagick.COLORSPACE_YCC {
		return "COLORSPACE_YCC"
	} else if colorSpace == imagick.COLORSPACE_YIQ {
		return "COLORSPACE_YIQ"
	} else if colorSpace == imagick.COLORSPACE_YPBPR {
		return "COLORSPACE_YPBPR"
	} else if colorSpace == imagick.COLORSPACE_YUV {
		return "COLORSPACE_YUV"
	} else if colorSpace == imagick.COLORSPACE_CMYK {
		return "COLORSPACE_CMYK"
	} else if colorSpace == imagick.COLORSPACE_SRGB {
		return "COLORSPACE_SRGB"
	} else if colorSpace == imagick.COLORSPACE_HSB {
		return "COLORSPACE_HSB"
	} else if colorSpace == imagick.COLORSPACE_HSL {
		return "COLORSPACE_HSL"
	} else if colorSpace == imagick.COLORSPACE_HWB {
		return "COLORSPACE_HWB"
	} else if colorSpace == imagick.COLORSPACE_REC601LUMA {
		return "COLORSPACE_REC601LUMA"
	} else if colorSpace == imagick.COLORSPACE_REC601YCBCR {
		return "COLORSPACE_REC601YCBCR"
	} else if colorSpace == imagick.COLORSPACE_REC709LUMA {
		return "COLORSPACE_REC709LUMA"
	} else if colorSpace == imagick.COLORSPACE_REC709YCBCR {
		return "COLORSPACE_REC709YCBCR"
	} else if colorSpace == imagick.COLORSPACE_LOG {
		return "COLORSPACE_LOG"
	} else if colorSpace == imagick.COLORSPACE_CMY {
		return "COLORSPACE_CMY"
	} else {
		return "Unknown"
	}
}

func resize(wand *imagick.MagickWand, wandLinear *imagick.MagickWand, size Size, quality Quality, target string) error {
	var err error
	var mw *imagick.MagickWand

	var colorSpace imagick.ColorspaceType

	if size.Width == 0 && size.Height == 0 {
		mw = wand.Clone()
		defer mw.Destroy()

		colorSpace = mw.GetImageColorspace()
		if colorSpace == imagick.COLORSPACE_UNDEFINED {
			colorSpace = imagick.COLORSPACE_SRGB
		}
	} else {
		mw = wandLinear.Clone()
		defer mw.Destroy()

		colorSpace = mw.GetImageColorspace()
		if colorSpace == imagick.COLORSPACE_UNDEFINED {
			colorSpace = imagick.COLORSPACE_SRGB
		}

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
	iccProfile := wand.GetImageProfile("ICC")
	if iccProfile != "" {
		mw.SetImageProfile("ICC", []byte(iccProfile))
	}
	iptcProfile := wand.GetImageProfile("IPTC")
	if iptcProfile != "" {
		mw.SetImageProfile("IPTC", []byte(iptcProfile))
	}

	println(colorSpaceName(colorSpace))
	err = mw.WriteImage(target)

	return err
}
