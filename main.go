package main

import (
	"bufio"
	"flag"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/jasonlvhit/gocron"
	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/facebot"
	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
	interval = flag.Duration("interval", time.Minute*30, "posting interval")
	minfaces = flag.Int("minfaces", 1, "minimum faces")
	upload   = flag.Bool("upload", false, "enable photo uploading")
	testimg  = flag.String("test.image", "", "test image")
	testdir  = flag.String("test.dir", "", "test a directory of images")
	facedir  = flag.String("face.dir", "faces", "directory to load faces from")
	postTime = flag.String("post.time", "19:00", "time of day to post")

	importLegacy = flag.String("import.legacy", "", "import a legacy database")
)

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
	log.Printf("found %d face(s) in image\n", len(faces))

	newImage, err := faceutil.DrawFaces(baseImage, faces)
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

func main() {
	flag.Parse()

	faceutil.MustLoadFaces(*facedir)

	if *importLegacy != "" {
		store, err := imgstore.Open("media.db")
		if err != nil {
			log.Fatal(err)
		}
		if err := store.ImportLegacyDatabase(*importLegacy); err != nil {
			log.Fatal(err)
		}
		return
	}

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

	captions, err := readCaptions("captions.txt")
	if err != nil {
		log.Fatal(err)
	}
	shuffle(captions)

	bot, err := facebot.New(&facebot.Options{
		Username: *username,
		Password: *password,
		MinFaces: *minfaces,
		Upload:   *upload,
		Captions: captions,
	})
	if err != nil {
		log.Fatal(err)
	}

	gocron.Every(1).Day().At(*postTime).Do(func() {
		if err := bot.Post(); err != nil {
			log.Printf("posting error: %s\n", err)
		}
	})

	gocron.Start()
	bot.Run()
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
