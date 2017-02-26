package main

import (
	"bufio"
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

	"github.com/jasonlvhit/gocron"
	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/facebot"
	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
)

var (
	username = flag.String("username", "", "instagram username")
	password = flag.String("password", "", "instagram password")
	minfaces = flag.Int("minfaces", 1, "minimum faces")
	upload   = flag.Bool("upload", false, "enable photo uploading")
	testimg  = flag.String("test.image", "", "test image")
	testdir  = flag.String("test.dir", "", "test a directory of images")
	facedir  = flag.String("face.dir", "faces", "directory to load faces from")
	httpport = flag.String("http.port", "", "http port (example :8080)")
	coinflip = flag.Bool("coinflip", true, "50/50 chance of posting at the requested times")

	resetStore = flag.Bool("reset.store", false, "mark all store records as available")
	storefile  = flag.String("store", "store.db", "the store file")

	postInterval = flag.Duration("post.interval", 0, "how often to post")
	postTimes    times
)

func init() {
	flag.Var(&postTimes, "post.time", "time to post")
}

type times []string

func (t *times) String() string {
	return fmt.Sprint(*t)
}

func (t *times) Set(value string) error {
	*t = append(*t, value)
	return nil
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
	log.Printf("found %d face(s) in image\n", len(faces))
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

func main() {
	flag.Parse()

	faceutil.MustLoadFaces(*facedir)

	store, err := imgstore.Open(*storefile)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if *resetStore {
		if err := store.ResetStates(); err != nil {
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
		Store:    store,
	})
	if err != nil {
		log.Fatal(err)
	}
	bot.Start()

	if *httpport != "" {
		http.HandleFunc("/demo", func(w http.ResponseWriter, r *http.Request) {
			img, err := bot.Demo()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Add("Content-Type", "image/jpeg")
			if err := jpeg.Encode(w, img, &jpeg.Options{Quality: jpeg.DefaultQuality}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		go func() {
			if err := http.ListenAndServe(*httpport, nil); err != nil {
				log.Printf("error: %s\n", err)
			}
		}()
	}

	maybeDoPost := func() {
		if *coinflip && rand.Intn(2) == 0 {
			return
		}
		if err := bot.Post(); err != nil {
			log.Printf("posting error: %s\n", err)
		}
	}

	switch {
	case *postInterval != 0:
		for {
			maybeDoPost()
			time.Sleep(*postInterval)
		}
	case len(postTimes) > 0:
		for _, t := range postTimes {
			gocron.Every(1).Day().At(t).Do(maybeDoPost)
		}
		<-gocron.Start()
	default:
		select {}
	}
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
