package faceutil

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
)

type ByCenterY []image.Rectangle

func (b ByCenterY) Len() int {
	return len(b)
}

func (b ByCenterY) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b ByCenterY) Less(i, j int) bool {
	var (
		p1 = getRectCenter(b[i])
		p2 = getRectCenter(b[j])
	)
	return p1.Y < p2.Y
}

func addRectPadding(pct float64, rect image.Rectangle, bounds image.Rectangle) image.Rectangle {
	var (
		width  = float64(rect.Dx())
		height = float64(rect.Dy())

		minScale    = 0.1
		widthScale  = math.Max(1.0-(height/float64(bounds.Dx()))-0.3, minScale)
		heightScale = math.Max(1.0-(width/float64(bounds.Dy()))-0.3, minScale)

		widthPadding  = int(widthScale * pct * (width / 100) / 2)
		heightPadding = int(heightScale * pct * (height / 100) / 2)
	)

	log.Printf(
		"rect: (%f x %f), bounds: (%d, %d), scale: (%f, %f), padding: (%d x %d)\n",
		width,
		height,
		bounds.Dx(),
		bounds.Dx(),
		widthScale,
		heightScale,
		widthPadding,
		heightPadding,
	)

	return image.Rect(
		rect.Min.X-widthPadding,
		rect.Min.Y-heightPadding*3,
		rect.Max.X+widthPadding,
		rect.Max.Y+heightPadding,
	)
}

func canvasFromImage(i image.Image) *image.NRGBA {
	bounds := i.Bounds()
	canvas := image.NewNRGBA(bounds)
	draw.Draw(canvas, bounds, i, bounds.Min, draw.Src)
	return canvas
}

func drawRect(img *image.NRGBA, rect image.Rectangle, c color.Color) {
	var (
		x1 = rect.Min.X
		x2 = rect.Max.X
		y1 = rect.Min.Y
		y2 = rect.Max.Y

		thickness = 2
	)
	for t := 0; t < thickness; t++ {
		// draw horizontal lines
		for x := x1; x <= x2; x++ {
			img.Set(x, y1+t, c)
			img.Set(x, y2-t, c)
		}
		// draw vertical lines
		for y := y1; y <= y2; y++ {
			img.Set(x1+t, y, c)
			img.Set(x2-t, y, c)
		}
	}
}

func getRectCenter(rect image.Rectangle) image.Point {
	return image.Point{
		X: rect.Min.X + rect.Dx()/2,
		Y: rect.Min.Y + rect.Dy()/2,
	}
}

func getRectCenteredIn(child, parent image.Rectangle) image.Rectangle {
	var (
		center = getRectCenter(parent)
		halfX  = child.Dx() / 2
		halfY  = child.Dy() / 2
	)
	return image.Rectangle{
		Min: image.Point{
			X: center.X - halfX,
			Y: center.Y - halfY,
		},
		Max: image.Point{
			X: center.X + halfX,
			Y: center.Y + halfY,
		},
	}
}
