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

type FaceReplacer struct {
	base  image.Image
	rects []image.Rectangle
	faces FaceList
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
		rects: finder.Detect(base),
		faces: faces,
		base:  base,
	}, nil
}

func (fr *FaceReplacer) NumFaces() int {
	return len(fr.rects)
}

func (fr *FaceReplacer) AddFaces() (*image.RGBA, error) {

	bounds := fr.base.Bounds()
	canvas := canvasFromImage(fr.base)

	for _, rect := range fr.rects {
		rect := rectMargin(30.0, rect)

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

	if len(fr.rects) == 0 {
		face := imaging.Resize(
			fr.faces[0],
			bounds.Dx()/3,
			0,
			imaging.Lanczos,
		)
		face_bounds := face.Bounds()
		draw.Draw(
			canvas,
			bounds,
			face,
			bounds.Min.Add(image.Pt(-bounds.Max.X/2+face_bounds.Max.X/2, -bounds.Max.Y+int(float64(face_bounds.Max.Y)/1.9))),
			draw.Over,
		)
	}

	return canvas, nil
}
