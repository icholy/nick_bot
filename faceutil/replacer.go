package faceutil

import (
	"flag"
	"image"
	"image/color"
	"sort"

	"github.com/disintegration/imaging"
	"github.com/lazywei/go-opencv/opencv"
)

var (
	minNeighboor = flag.Int("min.neighboor", 9, "the lower this number is, the more faces will be found")
	haarCascade  = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
	margin       = flag.Float64("margin", 60.0, "The face rectangle margin")
	faceOpacity  = flag.Float64("face.opacity", 1.0, "Face opacity [0-255]")

	shouldDrawFace  = flag.Bool("draw.face", true, "Draw the face")
	shouldDrawRects = flag.Bool("draw.rects", false, "Show the detection rectangles")
)

func DrawFace(canvas *image.NRGBA, faceRect image.Rectangle, primary bool) *image.NRGBA {
	var (
		// rect colors
		red   = color.RGBA{255, 0, 0, 255}
		green = color.RGBA{0, 255, 0, 255}
		blue  = color.RGBA{0, 0, 255, 255}

		// select a random source face
		srcFaceImg = randomFace(primary)

		// add padding around detected face rect
		paddedRect = addRectPadding(*margin, faceRect, canvas.Bounds())

		// resize the face image to fit inside the padded rect
		faceImg = imaging.Resize(srcFaceImg, paddedRect.Dx(), 0, imaging.Lanczos)

		// center the face rect size inside the padded rect
		placementRect = getRectCenteredIn(faceImg.Rect, paddedRect)
	)

	if *shouldDrawFace {
		canvas = imaging.Overlay(canvas, faceImg, placementRect.Min, *faceOpacity)
	}

	if *shouldDrawRects {
		drawRect(canvas, faceRect, red)
		drawRect(canvas, paddedRect, green)
		drawRect(canvas, placementRect, blue)
	}

	return canvas
}

func DrawFaces(base image.Image, rects []image.Rectangle) *image.NRGBA {
	var (
		canvas     = canvasFromImage(base)
		usePrimary = len(rects) < 4
	)
	for _, faceRect := range rects {
		canvas = DrawFace(canvas, faceRect, usePrimary)
	}
	return canvas
}

func DetectFaces(i image.Image) []image.Rectangle {
	var (
		output  []image.Rectangle
		cascade = opencv.LoadHaarClassifierCascade(*haarCascade)
		faces   = cascade.DetectObjects(opencv.FromImage(i), *minNeighboor)
	)
	defer cascade.Release()
	for _, face := range faces {
		output = append(output, image.Rectangle{
			Min: image.Point{face.X(), face.Y()},
			Max: image.Point{face.X() + face.Width(), face.Y() + face.Height()},
		})
	}
	sort.Sort(ByCenterY(output))
	return output
}

func ReplaceFaces(i image.Image) *image.NRGBA {
	faces := DetectFaces(i)
	return DrawFaces(i, faces)
}
