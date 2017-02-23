package facefinder

import (
	"flag"
	"image"

	"github.com/lazywei/go-opencv/opencv"
)

var minNeighboor = flag.Int("min.neighboor", 5, "the lower this number is, the more faces will be found")

var faceCascade *opencv.HaarCascade

type Finder struct {
	cascade *opencv.HaarCascade
}

func NewFinder(xml string) *Finder {
	return &Finder{
		cascade: opencv.LoadHaarClassifierCascade(xml),
	}
}

func (f *Finder) Close() {
	f.cascade.Release()
}

func (f *Finder) Detect(i image.Image) []image.Rectangle {
	var output []image.Rectangle

	faces := f.cascade.DetectObjects(opencv.FromImage(i), *minNeighboor)
	for _, face := range faces {
		output = append(output, image.Rectangle{
			image.Point{face.X(), face.Y()},
			image.Point{face.X() + face.Width(), face.Y() + face.Height()},
		})
	}

	return output
}
