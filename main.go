package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/replacer"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
	interval = flag.Duration("interval", time.Minute*30, "posting interval")
	minfaces = flag.Int("minfaces", 1, "minimum faces")
	upload   = flag.Bool("upload", false, "enable photo uploading")
	testimg  = flag.String("test.image", "", "test image")
	testdir  = flag.String("test.dir", "", "test a directory of images")
)

type Media struct {
	ID       string
	URL      string
	UserID   string
	Username string
	Image    image.Image
}

func (m *Media) String() string {
	return fmt.Sprintf("%s: %s", m.Username, m.URL)
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
	faceReplacer, err := replacer.New(baseImage, "faces")
	if err != nil {
		return err
	}

	log.Printf("found %d face(s) in image\n", faceReplacer.NumFaces())
	if faceReplacer.NumFaces() < *minfaces {
		return fmt.Errorf("not enough faces")
	}

	newImage, err := faceReplacer.AddFaces()
	if err != nil {
		return err
	}

	return jpeg.Encode(w, newImage, &jpeg.Options{jpeg.DefaultQuality})
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

func attempt(db *sql.DB, caption string) error {

	session, err := instagram.New(*username, *password)
	if err != nil {
		return err
	}
	defer session.Close()

	// get a list of users
	users, err := session.GetUsers()
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return fmt.Errorf("no users found")
	}
	log.Printf("found %d users\n", len(users))

	var media *Media

	for i := 0; true; i++ {

		if i > 10 {
			return fmt.Errorf("too many failed attempts")
		}

		log.Printf("fetch attempt %d\n", i+1)
		media, err = fetchRandomMedia(db, session, users)
		if err == nil {
			break
		}

		log.Printf("error fetching media: %s\n", err)
		log.Println("trying again in 5 seconds\n")
		time.Sleep(time.Second * 5)
	}

	faceReplacer, err := replacer.New(media.Image, "faces")
	if err != nil {
		return err
	}

	facecount := faceReplacer.NumFaces()

	if err := saveMedia(db, media, facecount); err != nil {
		return err
	}

	log.Printf("found %d face(s) in image\n", faceReplacer.NumFaces())
	if faceReplacer.NumFaces() < *minfaces {
		return fmt.Errorf("not enough faces")
	}

	newImage, err := faceReplacer.AddFaces()
	if err != nil {
		return err
	}

	outpath := filepath.Join("output", media.ID+"_nick.jpeg")
	if err := writeImage(outpath, newImage); err != nil {
		return err
	}

	log.Printf("written to %s\n", outpath)

	if *upload {
		caption := fmt.Sprintf("%s\n\nphotocred goes to: @%s", caption, media.Username)
		return session.UploadPhoto(outpath, caption)
	}

	return nil
}

func main() {
	flag.Parse()

	if *testimg != "" {
		if err := testImage(*testimg, os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *testdir != "" {
		if err := testImageDir(*testdir); err != nil {
			log.Fatal(err)
		}
		return
	}

	captionIndex := 0
	captions, err := readCaptions("captions.txt")
	if err != nil {
		log.Fatal(err)
	}
	shuffle(captions)

	db, err := sql.Open("sqlite3", "media.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createDatabase(db); err != nil {
		log.Fatal(err)
	}

	for {

		caption := captions[captionIndex]
		captionIndex++
		if captionIndex >= len(captions) {
			captionIndex = 0
		}

		log.Println("trying to post an image")
		if err := attempt(db, caption); err != nil {
			log.Printf("error: %s\n", err)
		}
		time.Sleep(*interval)
	}

}

func writeImage(filename string, img image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
}

func fetchRandomMedia(db *sql.DB, session *instagram.Session, users []*instagram.User) (*Media, error) {

	// select a random user
	user := users[rand.Intn(len(users))]
	log.Printf("Randomly selected user: %s\n", user.Name)

	medias, err := session.GetRecentUserMedias(user)
	if err != nil {
		return nil, err
	}
	log.Printf("Got %d medias\n", len(medias))

	// find unused medias
	var unused []*instagram.Media
	for _, media := range medias {
		ok, err := hasMediaID(db, media.ID)
		if err != nil {
			return nil, err
		}
		if !ok {
			unused = append(unused, media)
		}
	}
	if len(unused) == 0 {
		return nil, fmt.Errorf("no unused images for user")
	}

	// select a random media
	media := unused[rand.Intn(len(unused))]

	// get the image
	resp, err := http.Get(media.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Media{
		ID:       media.ID,
		URL:      media.URL,
		UserID:   user.ID,
		Username: user.Name,
		Image:    img,
	}, nil
}

func createDatabase(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			user_id,
			user_name,
			media_id,
			media_url,
			face_count
		)
	`)
	return err
}

func saveMedia(db *sql.DB, media *Media, facecount int) error {
	log.Printf("Saving: %s\n", media)
	_, err := db.Exec(
		`INSERT INTO media VALUES (?, ?, ?, ?, ?)`,
		media.UserID, media.Username, media.ID, media.URL, facecount,
	)
	return err
}

func hasMediaID(db *sql.DB, mediaID string) (bool, error) {
	var count int
	if err := db.QueryRow(
		`SELECT COUNT(1) FROM media WHERE media_id = ? LIMIT 1`,
		mediaID,
	).Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func shuffle(slice []string) {
	for i := range slice {
		j := rand.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}

func readCaptions(filename string) ([]string, error) {
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
