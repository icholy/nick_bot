package replacer

import (
	"image"
	"image/color"
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

func drawRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
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
