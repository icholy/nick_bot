package replacer

import (
	"image"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func loadImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return m, nil
}

type FaceList []image.Image

func (fl FaceList) Random() image.Image {
	i := rand.Intn(len(fl))
	face := fl[i]
	if rand.Intn(2) == 0 {
		return imaging.FlipH(face)
	}
	return face
}

func LoadFaces(dir string) (FaceList, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var faces FaceList
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".png" {
			continue
		}
		m, err := loadImage(path.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}
		faces = append(faces, m)
	}
	return faces, nil
}
