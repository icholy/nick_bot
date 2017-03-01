package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/facebot"
	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
	"github.com/icholy/nick_bot/model"
)

var (
	username   = flag.String("username", "", "instagram username")
	password   = flag.String("password", "", "instagram password")
	minfaces   = flag.Int("minfaces", 1, "minimum faces")
	upload     = flag.Bool("upload", false, "enable photo uploading")
	testimg    = flag.String("test.image", "", "test image")
	testdir    = flag.String("test.dir", "", "test a directory of images")
	facedir    = flag.String("face.dir", "faces", "directory to load faces from")
	httpport   = flag.String("http.port", "", "http port (example :8080)")
	autofollow = flag.Bool("autofollow", false, "auto follow random people")

	resetStore = flag.Bool("reset.store", false, "mark all store records as available")
	storefile  = flag.String("store", "store.db", "the store file")

	postNow      = flag.Bool("post.now", false, "post and exit")
	postInterval = flag.Duration("post.interval", 0, "how often to post")
)

var banner = `
  _  _ _    _     ___      _
 | \| (_)__| |__ | _ ) ___| |
 | .' | / _| / / | _ \/ _ \  _|
 |_|\_|_\__|_\_\ |___/\___/\__|

 Adding some much needed nick to your photos.
`

func main() {
	flag.Parse()

	faceutil.MustLoadFaces(*facedir)

	store, err := imgstore.Open(*storefile)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	switch {
	case *resetStore:
		if err := store.ResetStates(); err != nil {
			log.Fatal(err)
		}
	case *testimg != "":
		if err := testImage(*testimg, os.Stdout); err != nil {
			log.Fatal(err)
		}
	case *testdir != "":
		if err := testImageDir(*testdir); err != nil {
			log.Fatal(err)
		}
	default:
		if err := startBot(store); err != nil {
			log.Fatal(err)
		}
	}
}

func startBot(store *imgstore.Store) error {

	fmt.Println(banner)

	captions, err := readLines("captions.txt")
	if err != nil {
		return err
	}
	shuffle(captions)

	bot, err := facebot.New(&facebot.Options{
		Username:   *username,
		Password:   *password,
		MinFaces:   *minfaces,
		Upload:     *upload,
		AutoFollow: *autofollow,
		Captions:   captions,
		Store:      store,
	})
	if err != nil {
		return err
	}
	go bot.Run()

	if *httpport != "" {
		go runHTTPServer(bot, store)
	}

	doPost := func() {
		if err := bot.Post(); err != nil {
			log.Printf("posting error: %s\n", err)
		}
	}

	switch {
	case *postNow:
		doPost()
		return nil
	case *postInterval != 0:
		for {
			doPost()
			time.Sleep(*postInterval)
		}
	default:
		schedule, err := loadSchedule("schedule.json")
		if err != nil {
			log.Fatal(err)
		}
		for day, times := range schedule {
			day = strings.ToLower(day)
			for _, t := range times {
				switch day {
				case "monday":
					gocron.Every(1).Monday().At(t).Do(doPost)
				case "tuesday":
					gocron.Every(1).Tuesday().At(t).Do(doPost)
				case "wednesday":
					gocron.Every(1).Wednesday().At(t).Do(doPost)
				case "thursday":
					gocron.Every(1).Thursday().At(t).Do(doPost)
				case "friday":
					gocron.Every(1).Friday().At(t).Do(doPost)
				case "saturday":
					gocron.Every(1).Saturday().At(t).Do(doPost)
				case "sunday":
					gocron.Every(1).Sunday().At(t).Do(doPost)
				case "everyday":
					gocron.Every(1).Day().At(t).Do(doPost)
				default:
					log.Fatalf("invalid schedule key: %s", day)
				}
			}
		}
		<-gocron.Start()
	}

	return nil
}

func runHTTPServer(bot *facebot.Bot, store *imgstore.Store) {
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
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		stats, err := store.Stats(model.MediaAvailable)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	if err := http.ListenAndServe(*httpport, nil); err != nil {
		log.Printf("error: %s\n", err)
	}
}
