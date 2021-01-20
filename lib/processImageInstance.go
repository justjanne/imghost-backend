package lib

import (
	"gopkg.in/gographics/imagick.v2/imagick"
	"strings"
)

func ProcessImageInstance(
	wand *imagick.MagickWand,
	size Size,
	quality Quality,
	parameters ImageParameters,
	target string,
) error {
	var err error
	var mw *imagick.MagickWand

	mw = wand.Clone()
	defer mw.Destroy()

	if err = ResizeImageInstance(mw, size); err != nil {
		return err
	}
	for _, key := range mw.GetImageProperties("*") {
		if !strings.HasPrefix(key, "png:") {
			if err = mw.DeleteImageProperty(key); err != nil {
				return err
			}
		}
	}

	if iccProfile, iccProfileOk := parameters.profiles["icc"]; iccProfileOk {
		err = mw.ProfileImage("icc", iccProfile)
	} else if iptcProfile, iptcProfileOk := parameters.profiles["iptc"]; iptcProfileOk {
		err = mw.ProfileImage("iptc", iptcProfile)
	} else {
		err = mw.TransformImageColorspace(parameters.colorspace)
	}

	if err = mw.SetImageDepth(parameters.depth); err != nil {
		return err
	}

	if quality.CompressionQuality != 0 {
		_ = mw.SetImageCompressionQuality(quality.CompressionQuality)
	}

	if len(quality.SamplingFactors) != 0 {
		_ = mw.SetSamplingFactors(quality.SamplingFactors)
	}

	err = mw.WriteImage(target)

	return err
}
