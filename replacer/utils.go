package replacer

import (
	"image"
	"image/draw"
)

func rectMargin(pct float64, rect image.Rectangle) image.Rectangle {
	width := float64(rect.Max.X - rect.Min.X)
	height := float64(rect.Max.Y - rect.Min.Y)

	padding_width := int(pct * (width / 100) / 2)
	padding_height := int(pct * (height / 100) / 2)

	return image.Rect(
		rect.Min.X-padding_width,
		rect.Min.Y-padding_height*3,
		rect.Max.X+padding_width,
		rect.Max.Y+padding_height,
	)
}

func canvasFromImage(i image.Image) *image.RGBA {
	bounds := i.Bounds()
	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, i, bounds.Min, draw.Src)

	return canvas
}
