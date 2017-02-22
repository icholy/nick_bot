package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/replacer"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
)

func main() {
	flag.Parse()

	session, err := instagram.New(*username, *password)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	base, mediaID, err := fetchRandomImage(session)
	if err != nil {
		log.Fatal(err)
	}

	newImage, err := replaceFaces(base)
	if err != nil {
		log.Fatal(err)
	}

	outpath := filepath.Join("output", mediaID+"_nick.jpeg")
	if err := writeImage(outpath, newImage); err != nil {
		log.Fatal(err)
	}

	log.Printf("written to %s\n", outpath)
}

func writeImage(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
}

func fetchRandomImage(session *instagram.Session) (image.Image, string, error) {

	// get a list of users
	users, err := session.GetUsers()
	if err != nil {
		return nil, "", err
	}
	if len(users) == 0 {
		return nil, "", fmt.Errorf("no users found")
	}
	log.Printf("found %d users\n", len(users))

	// select a random user
	user := users[rand.Intn(len(users))]
	log.Printf("randomly selected: %s\n", user.Name)

	// get a list of media ids for the user
	medias, err := session.GetUserMediaIDS(user)
	if err != nil {
		return nil, "", err
	}
	if len(medias) == 0 {
		return nil, "", fmt.Errorf("no medias found for user")
	}
	log.Printf("found %d media ids\n", len(medias))

	// select a random media id
	mediaID := medias[rand.Intn(len(medias))]
	log.Printf("randomly selected id %s\n", mediaID)

	// get the url
	media, err := session.GetUserImage(mediaID)
	if err != nil {
		return nil, "", err
	}
	log.Printf("got url for media id: %s\n", media.URL)

	// get the image
	resp, err := http.Get(media.URL)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	log.Printf("fetched the image\n")

	// decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return img, mediaID, nil
}

func replaceFaces(base image.Image) (image.Image, error) {
	faceReplacer, err := replacer.New(base, "faces")
	if err != nil {
		return nil, err
	}
	if faceReplacer.NumFaces() == 0 {
		return nil, fmt.Errorf("no faces found")
	}
	log.Printf("found %d face(s) in image\n", faceReplacer.NumFaces())
	return faceReplacer.AddFaces()
}
