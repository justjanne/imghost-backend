package lib

import (
	"errors"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"math"
)

func ResizeImageInstance(mw *imagick.MagickWand, size Size) error {
	var err error
	if size.Width != 0 || size.Height != 0 {
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

				if err = mw.CropImage(cWidth, cHeight, dx, dy); err != nil {
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
			if err = mw.ResizeImage(nWidth, nHeight, imagick.FILTER_BOX, 1); err != nil {
				return err
			}
		}
	}

	return nil
}
