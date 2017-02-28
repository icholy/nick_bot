package faceutil

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

func addRectPadding(pct float64, rect image.Rectangle, bounds image.Rectangle) image.Rectangle {
	var (
		width  = float64(rect.Dx())
		height = float64(rect.Dy())

		boundsWidth  = float64(bounds.Dx())
		boundsHeight = float64(bounds.Dx())

		widthScale  = (1.0 / (width / boundsWidth)) / boundsWidth
		heightScale = (1.0 / (height / boundsHeight)) / boundsHeight
		scale       = math.Min(widthScale, heightScale)

		widthPadding  = int(scale * pct * (width / 100) / 2)
		heightPadding = int(scale * pct * (height / 100) / 2)
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

func getRectCenteredIn(child, parent image.Rectangle) image.Rectangle {
	center := image.Point{
		X: parent.Min.X + parent.Dx()/2,
		Y: parent.Min.Y + parent.Dy()/2,
	}
	halfX := child.Dx() / 2
	halfY := child.Dy() / 2
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
