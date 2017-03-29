package faceutil

import (
	"image"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/disintegration/imaging"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	primaryFaceList []image.Image
	allFaceList     []image.Image
)

func LoadFaces(dir string) error {
	var err error
	primaryFaceList, err = loadFaces(filepath.Join(dir, "primary"))
	if err != nil {
		return err
	}
	secondayFaceList, err := loadFaces(filepath.Join(dir, "seconday"))
	if err != nil {
		return err
	}
	for _, face := range primaryFaceList {
		allFaceList = append(allFaceList, face)
	}
	for _, face := range secondayFaceList {
		allFaceList = append(allFaceList, face)
	}
	return nil
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

func randomFace(primary bool) image.Image {
	var faces []image.Image
	if primary {
		faces = primaryFaceList
	} else {
		faces = allFaceList
	}
	i := rand.Intn(len(faces))
	face := faces[i]
	if rand.Intn(2) == 0 {
		return imaging.FlipH(face)
	}
	return face
}
