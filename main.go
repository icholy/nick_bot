package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/replacer"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
)

func try(session *instagram.Session, users []*instagram.User) error {

	// select a random user
	user := users[rand.Intn(len(users))]
	log.Printf("randomly selected: %s\n", user.Name)

	// get a list of media ids for the user
	medias, err := session.GetUserMediaIDS(user)
	if err != nil {
		return err
	}
	if len(medias) == 0 {
		return fmt.Errorf("no medias found for user")
	}
	log.Printf("found %d media ids\n", len(medias))

	// select a random media id
	mediaID := medias[rand.Intn(len(medias))]
	log.Printf("randomly selected id %s\n", mediaID)

	// get the url
	image, err := session.GetUserImage(mediaID)
	if err != nil {
		return err
	}
	log.Printf("got url for media id: %s\n", image.URL)

	// get the image
	resp, err := http.Get(image.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Printf("fetched the image\n")

	faceReplacer, err := replacer.New(resp.Body, "faces")
	if err != nil {
		return err
	}

	if faceReplacer.NumFaces() == 0 {
		return fmt.Errorf("no faces found")
	}
	log.Printf("found %d face(s) in image\n", faceReplacer.NumFaces())

	img, err := faceReplacer.AddFaces()
	if err != nil {
		log.Fatal(err)
	}

	outpath := filepath.Join("output", mediaID+".jpeg")
	f, err := os.Create(outpath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality}); err != nil {
		log.Fatal(err)
	}

	log.Printf("written to %s\n", outpath)
	return nil
}

func main() {
	flag.Parse()

	// login
	session, err := instagram.New(*username, *password)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// get a list of users
	users, err := session.GetUsers()
	if err != nil {
		log.Fatal(err)
	}
	if len(users) == 0 {
		log.Fatalf("no users found")
	}
	log.Printf("found %d users\n", len(users))

	// keep trying
	for i := 0; i < 10; i++ {
		log.Printf("attempt: %d\n", i)
		err := try(session, users)
		if err == nil {
			break
		}
		log.Printf("error: %s\n", err)
		time.Sleep(time.Second * 30)
	}
}
