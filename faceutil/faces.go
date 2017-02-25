package faceutil

import (
	"image"
	"io/ioutil"
	"log"
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

var faceList []image.Image

func LoadFaces(dir string) error {
	var err error
	faceList, err = loadFaces(dir)
	return err
}

func MustLoadFaces(dir string) {
	if err := LoadFaces(dir); err != nil {
		log.Fatal(err)
	}
}

func loadFaces(dir string) ([]image.Image, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var faces []image.Image
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

func randomFace() image.Image {
	i := rand.Intn(len(faceList))
	face := faceList[i]
	if rand.Intn(2) == 0 {
		return imaging.FlipH(face)
	}
	return face
}
