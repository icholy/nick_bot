package replacer

import (
	"errors"
	"flag"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"github.com/lazywei/go-opencv/opencv"
)

var minNeighboor = flag.Int("min.neighboor", 5, "the lower this number is, the more faces will be found")
var haarCascade = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
var margin = flag.Float64("margin", 50.0, "The face rectangle margin")
var showRects = flag.Bool("show.rects", false, "Show the detection rectangles")
var drawFace = flag.Bool("draw.face", true, "Draw the face")
var faceOpacity = flag.Float64("face.opacity", 1.0, "Face opacity [0-255]")

var showRectDetection = flag.Bool("show.rects.detection", true, "Show the detection rects")
var showRectPadding = flag.Bool("show.rects.padding", true, "Show the padding rects")
var showRectPlacement = flag.Bool("show.rects.placement", true, "Show the placement rects")

type Replacer struct {
	base  image.Image
	rects []image.Rectangle
	faces FaceList
}

func New(base image.Image, facesPath string) (*Replacer, error) {

	// read faces
	faces, err := loadFaces(facesPath)
	if err != nil {
		return nil, err
	}
	if len(faces) == 0 {
		return nil, errors.New("no faces found")
	}

	// find faces in base image
	return &Replacer{
		rects: DetectFaces(base),
		faces: faces,
		base:  base,
	}, nil
}

func (rep *Replacer) NumFaces() int {
	return len(rep.rects)
}

func (rep *Replacer) AddFaces() (*image.NRGBA, error) {

	var (
		canvas = canvasFromImage(rep.base)

		red   = color.RGBA{255, 0, 0, 255}
		green = color.RGBA{0, 255, 0, 255}
		blue  = color.RGBA{0, 0, 255, 255}
	)

	for _, faceRect := range rep.rects {

		srcFaceImg := rep.faces.Random()

		// add padding around detected face rect
		paddingRect := addRectPadding(*margin, faceRect)

		// resize the face image to fit inside the padded rect
		faceImg := imaging.Fit(srcFaceImg, paddingRect.Dx(), paddingRect.Dy(), imaging.Lanczos)

		// center the face rect size inside the padded rect
		placementRect := getRectCenteredIn(faceImg.Rect, paddingRect)

		// draw the face
		if *drawFace {
			canvas = imaging.Overlay(canvas, faceImg, placementRect.Min, *faceOpacity)
		}

		if *showRects {
			if *showRectDetection {
				drawRect(canvas, faceRect, red)
			}
			if *showRectPadding {
				drawRect(canvas, paddingRect, green)
			}
			if *showRectPlacement {
				drawRect(canvas, placementRect, blue)
			}
		}
	}

	return canvas, nil
}

func DetectFaces(i image.Image) []image.Rectangle {
	var (
		output  []image.Rectangle
		cascade = opencv.LoadHaarClassifierCascade(*haarCascade)
	)
	defer cascade.Release()
	faces := cascade.DetectObjects(opencv.FromImage(i), *minNeighboor)
	for _, face := range faces {
		output = append(output, image.Rectangle{
			image.Point{face.X(), face.Y()},
			image.Point{face.X() + face.Width(), face.Y() + face.Height()},
		})
	}
	return output
}
