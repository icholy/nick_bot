package bot

import (
	"image"
	"image/jpeg"
	"os"
)

func writeImage(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
}
