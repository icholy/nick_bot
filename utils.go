package main

import (
	"bufio"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/icholy/nick_bot/faceutil"
)

func shuffle(slice []string) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var captions []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		captions = append(captions, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return captions, err
}

func testImage(imgfile string, w io.Writer) error {
	f, err := os.Open(imgfile)
	if err != nil {
		return err
	}
	defer f.Close()
	baseImage, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	faces := faceutil.DetectFaces(baseImage)
	log.Infof("found %d face(s) in image\n", len(faces))
	newImage := faceutil.DrawFaces(baseImage, faces)
	return jpeg.Encode(w, newImage, &jpeg.Options{Quality: jpeg.DefaultQuality})
}

func testImageDir(dir string) error {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		var (
			srcFile = filepath.Join(dir, e.Name())
			dstFile = filepath.Join(dir, "nick_"+e.Name())
		)
		f, err := os.Create(dstFile)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := testImage(srcFile, f); err != nil {
			return err
		}
	}
	return nil
}
