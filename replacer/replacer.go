package replacer

import (
	"errors"
	"flag"
	"image"
	"image/draw"

	"github.com/disintegration/imaging"

	"github.com/icholy/nick_bot/replacer/facefinder"
)

var haarCascade = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
var margin = flag.Float64("margin", 30.0, "The face rectangle margin")

type FaceReplacer struct {
	base   image.Image
	rects  []image.Rectangle
	faces  FaceList
	finder *facefinder.Finder
}

func New(base image.Image, facesPath string) (*FaceReplacer, error) {

	// read faces
	var faces FaceList
	if err := faces.Load(facesPath); err != nil {
		return nil, err
	}
	if len(faces) == 0 {
		return nil, errors.New("no faces found")
	}

	// find faces in base image
	finder := facefinder.NewFinder(*haarCascade)

	return &FaceReplacer{
		rects:  finder.Detect(base),
		faces:  faces,
		base:   base,
		finder: finder,
	}, nil
}

func (fr *FaceReplacer) Close() {
	fr.finder.Close()
}

func (fr *FaceReplacer) NumFaces() int {
	return len(fr.rects)
}

func (fr *FaceReplacer) AddFaces() (*image.RGBA, error) {

	bounds := fr.base.Bounds()
	canvas := canvasFromImage(fr.base)

	for _, rect := range fr.rects {
		rect := rectMargin(*margin, rect)

		newFace := fr.faces.Random()
		if newFace == nil {
			panic("nil face")
		}
		face := imaging.Fit(newFace, rect.Dx(), rect.Dy(), imaging.Lanczos)

		draw.Draw(
			canvas,
			rect,
			face,
			bounds.Min,
			draw.Over,
		)
	}

	return canvas, nil
}
