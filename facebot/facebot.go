package facebot

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/model"
)

type Options struct {
	Username     string
	Password     string
	MinFaces     int
	Upload       bool
	PostInterval time.Duration
	Captions     []string
}

type Bot struct {
	opt   *Options
	store *imgstore.Store

	captionIndex int
}

func New(o *Options) (*Bot, error) {
	if o.MinFaces == 0 {
		o.MinFaces = 1
	}
	if o.PostInterval == 0 {
		o.PostInterval = time.Hour * 24
	}
	store, err := imgstore.Open("media.db")
	if err != nil {
		return nil, err
	}
	return &Bot{
		opt:   o,
		store: store,
	}, nil
}

func (b *Bot) getCaption(rec *model.Record) string {
	credit := fmt.Sprint("photocred goes to: @%s", rec.Username)
	captions := b.opt.Captions
	if len(captions) == 0 {
		return credit
	}
	caption := captions[b.captionIndex]
	b.captionIndex++
	if b.captionIndex >= len(captions) {
		b.captionIndex = 0
	}
	return fmt.Sprintf("%s\n\n%s", caption, credit)
}

func (b *Bot) Run() {

	// stat printer
	go func() {
		for {
			b.printStats()
			time.Sleep(time.Second * 30)
		}
	}()

	// crawler loop
	go func() {
		crawler := instagram.NewCrawler(b.opt.Username, b.opt.Password)
		for media := range crawler.Media() {
			log.Printf("bot: crawler found: %s\n", media)
			if err := b.handleMedia(media); err != nil {
				log.Printf("bot: %s\n", err)
			}
			time.Sleep(time.Second * 10)
		}
	}()

	// posting loop
	for {
		log.Println("bot: trying to post")
		if err := b.post(); err != nil {
			log.Printf("bot: %s\n", err)
		}
		log.Printf("bot: sleeping for %s\n", b.opt.PostInterval)
		time.Sleep(b.opt.PostInterval)
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

func (b *Bot) post() error {

	// find the best image
	rec, err := b.store.Search(b.opt.MinFaces)
	if err != nil {
		return err
	}
	log.Printf("bot: posting %s\n", rec)

	// try to post it
	if err := b.postRecord(rec); err != nil {
		log.Println("bot: %s", err)
		return b.store.SetState(rec.ID, model.MediaRejected)
	} else {
		return b.store.SetState(rec.ID, model.MediaUsed)
	}
}

func (b *Bot) postRecord(rec *model.Record) error {

	// download image
	img, err := fetchImage(rec.URL)
	if err != nil {
		return err
	}

	// replace the faces
	faces := faceutil.DetectFaces(img)
	newImage, err := faceutil.DrawFaces(img, faces)
	if err != nil {
		return err
	}

	// save image
	imgpath := filepath.Join("output", rec.ID+".jpeg")
	log.Printf("bot: writing image %s\n", imgpath)
	if err := writeImage(imgpath, newImage); err != nil {
		return err
	}

	if !b.opt.Upload {
		return nil
	}

	// upload photo
	log.Println("bot: uploading photo")
	session, err := instagram.NewSession(b.opt.Username, b.opt.Password)
	if err != nil {
		return err
	}
	defer session.Close()
	caption := b.getCaption(rec)
	return session.UploadPhoto(imgpath, caption)
}

func (b *Bot) printStats() {
	best, err := b.store.Search(0)
	if err != nil {
		log.Printf("bot: %s\n", err)
	} else {
		log.Printf("bot: best available: %s\n", best)
	}
	stats, err := b.store.Stats(model.MediaAvailable)
	if err != nil {
		log.Printf("bot: %s\n", err)
	} else if len(stats) == 0 {
		log.Println("bot: store stats: no data")
	} else {
		log.Printf("bot: store stats:\n%s\n", stats)
	}
}
