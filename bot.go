package main

import (
	"image"
	_ "image/png"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/faceutil"
	"github.com/icholy/nick_bot/imgstore"
	"github.com/icholy/nick_bot/instagram"
	"github.com/icholy/nick_bot/model"
)

type Bot struct {
	crawler *instagram.Crawler
	store   *imgstore.Store
}

func NewBot() (*Bot, error) {
	store, err := imgstore.Open("media.db")
	if err != nil {
		return nil, err
	}
	crawler, err := instagram.NewCrawler(*username, *password)
	if err != nil {
		return nil, err
	}
	bot := &Bot{crawler, store}
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

	// fetch the image
	resp, err := http.Get(m.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// detect the number of faces
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}
	faces := faceutil.DetectFaces(img)

	// write to store
	return b.store.Put(&model.Record{
		Media:     *m,
		FaceCount: len(faces),
		State:     model.MediaAvailable,
	})
}
