package replacer

import (
	"errors"
	"flag"
	"image"
	"image/color"

	"github.com/disintegration/imaging"

	"github.com/icholy/nick_bot/replacer/facefinder"
)

var haarCascade = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
var margin = flag.Float64("margin", 50.0, "The face rectangle margin")
var showRects = flag.Bool("show.rects", false, "Show the detection rectangles")
var faceOpacity = flag.Float64("face.opacity", 1.0, "Face opacity [0-255]")

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

func (fr *FaceReplacer) AddFaces() (*image.NRGBA, error) {

	var (
		canvas = canvasFromImage(fr.base)

		red   = color.RGBA{255, 0, 0, 255}
		green = color.RGBA{0, 255, 0, 255}
		blue  = color.RGBA{0, 0, 255, 255}
	)

	for _, rect := range fr.rects {

		srcFaceImg := fr.faces.Random()

		// add padding around detected face rect
		paddedRect := addRectPadding(*margin, rect)

		// resize the face image to fit inside the padded rect
		faceImg := imaging.Fit(srcFaceImg, paddedRect.Dx(), paddedRect.Dy(), imaging.Lanczos)

		// center the face rect size inside the padded rect
		faceRect := getRectCenteredIn(faceImg.Rect, paddedRect)

		// draw the face
		canvas = imaging.Overlay(canvas, faceImg, faceRect.Min, *faceOpacity)

		if *showRects {
			drawRect(canvas, rect, red)
			drawRect(canvas, paddedRect, green)
			drawRect(canvas, faceRect, blue)
		}
	}

	return canvas, nil
}
