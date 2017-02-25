package bot

import (
	"log"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/model"
)

type Bot struct {
	opt     *Options
	crawler *instagram.Crawler
	store   *imgstore.Store
}

type Options struct {
	Username string
	Password string
	MinFaces int
	Upload   bool
}

func NewBot(o *Options) (*Bot, error) {
	store, err := imgstore.Open("media.db")
	if err != nil {
		return nil, err
	}
	crawler, err := instagram.NewCrawler(o.Username, o.Password)
	if err != nil {
		return nil, err
	}
	bot := &Bot{
		opt:     o,
		crawler: crawler,
		store:   store,
	}
	go bot.start()
	return bot, nil
}

func (b *Bot) Stop() error {
	b.crawler.Stop()
	return b.store.Close()
}

func (b *Bot) start() {
	for media := range b.crawler.Media() {
		if err := b.handleMedia(media); err != nil {
			log.Printf("error: %s\n", err)
		}
		time.Sleep(time.Second * 10)
	}
}

func (b *Bot) handleMedia(m *model.Media) error {

	// make sure we haven't already processed this one
	skip, err := b.store.Has(m.ID)
	if skip || err != nil {
		return err
	}

	// download image
	img, err := fetchImage(m.URL)
	if err != nil {
		return err
	}

	// find the faces
	faces := faceutil.DetectFaces(img)

	// write to store
	return b.store.Put(&model.Record{
		Media:     *m,
		FaceCount: len(faces),
		State:     model.MediaAvailable,
	})
}

func (b *Bot) postImage() error {

	// find the best image
	rec, err := b.store.Search(b.opt.MinFaces)
	if err != nil {
		return err
	}

	// download image
	img, err := fetchImage(rec.URL)
	if err != nil {
		return err
	}

	// find the faces
	faces := faceutil.DetectFaces(img)

	// replace the faces
	newImage, err := faceutil.DrawFaces(img, faces)
	if err != nil {
		return err
	}

	// save image
	imgpath := filepath.Join("output", rec.ID+".jpeg")
	log.Printf("writing to %s\n", imgpath)
	if err := writeImage(imgpath, newImage); err != nil {
		return err
	}

	// login to instagram
	session, err := instagram.NewSession(b.opt.Username, b.opt.Password)
	if err != nil {
		return err
	}
	defer session.Close()

	// upload photo
	if b.opt.Upload {
		if err := session.UploadPhoto(imgpath, ""); err != nil {
			return err
		}
	}

	// update record state
	return b.store.SetState(rec.ID, model.MediaUsed)
}
