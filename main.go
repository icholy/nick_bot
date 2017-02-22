package main

import (
	"database/sql"
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

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/replacer"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
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

func main() {
	flag.Parse()

	db, err := sql.Open("sqlite3", "media.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := createDatabase(db); err != nil {
		log.Fatal(err)
	}

	session, err := instagram.New(*username, *password)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	media, err := fetchRandomMedia(db, session)
	if err != nil {
		log.Fatal(err)
	}

	faceReplacer, err := replacer.New(media.Image, "faces")
	if err != nil {
		log.Fatal(err)
	}

	facecount := faceReplacer.NumFaces()

	if err := saveMedia(db, media, facecount); err != nil {
		log.Fatal(err)
	}

	if faceReplacer.NumFaces() == 0 {
		log.Fatalf("no faces found")
	}
	log.Printf("found %d face(s) in image\n", faceReplacer.NumFaces())

	newImage, err := faceReplacer.AddFaces()
	if err != nil {
		log.Fatal(err)
	}

	outpath := filepath.Join("output", media.ID+"_nick.jpeg")
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

func fetchRandomMedia(db *sql.DB, session *instagram.Session) (*Media, error) {

	// get a list of users
	users, err := session.GetUsers()
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("no users found")
	}
	log.Printf("found %d users\n", len(users))

	// select a random user
	user := users[rand.Intn(len(users))]
	log.Printf("randomly selected: %s\n", user.Name)

	// get a list of media ids for the user
	medias, err := session.GetUserMediaIDS(user)
	if err != nil {
		return nil, err
	}
	log.Printf("found %d media ids\n", len(medias))

	// find an unused media id
	var mediaID string

	shuffle(medias)
	for _, id := range medias {
		ok, err := hasMediaID(db, id)
		if err != nil {
			return nil, err
		}
		if !ok {
			mediaID = id
			break
		}
	}
	if mediaID == "" {
		return nil, fmt.Errorf("no unused media found")
	}
	log.Printf("selected media id: %s\n", mediaID)

	// get the url
	media, err := session.GetUserImage(mediaID)
	if err != nil {
		return nil, err
	}
	log.Printf("got url for media id: %s\n", media.URL)

	// get the image
	resp, err := http.Get(media.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("fetched the image\n")

	// decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Media{
		ID:       mediaID,
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
