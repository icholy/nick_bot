package instagram

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/icholy/nick_bot/model"
)

var ErrStop = errors.New("stop crawler")

type Crawler struct {
	username string
	password string
	interval time.Duration

	usersCache   []*model.User
	usersUpdated time.Time

	out  chan *model.Media
	stop chan struct{}
}

func NewCrawler(username, password string) (*Crawler, error) {
	c := &Crawler{
		username: username,
		password: password,
		interval: 10 * time.Minute,

		out:  make(chan *model.Media),
		stop: make(chan struct{}),
	}
	go c.loop()
	return c, nil
}

func (c *Crawler) Media() <-chan *model.Media {
	return c.out
}

func (c *Crawler) loop() {
	for {
		select {
		case <-time.After(c.interval):
			err := c.crawl()
			switch {
			case err == ErrStop:
				return
			case err != nil:
				log.Printf("error: %s\n", err)
			}
		case <-c.stop:
			return
		}
	}
}

func (c *Crawler) crawl() error {
	s, err := NewSession(c.username, c.password)
	if err != nil {
		return err
	}
	defer s.Close()
	user, err := c.getRandomUser(s)
	if err != nil {
		return err
	}
	medias, err := s.GetRecentUserMedias(user)
	if err != nil {
		return err
	}
	for _, media := range medias {
		select {
		case c.out <- media:
			return nil
		case <-c.stop:
			return ErrStop
		}
	}
	return nil
}

func (c *Crawler) getRandomUser(s *Session) (*model.User, error) {

	// update the user cache if it's been over an hour
	if c.usersUpdated.IsZero() || time.Since(c.usersUpdated) > time.Hour {
		users, err := s.GetUsers()
		if err != nil {
			return nil, err
		}
		c.usersCache = users
		c.usersUpdated = time.Now()
	}

	// select a random user
	if len(c.usersCache) == 0 {
		return nil, fmt.Errorf("no users found")
	}
	index := rand.Intn(len(c.usersCache))
	return c.usersCache[index], nil
}

func (c *Crawler) Stop() {
	close(c.stop)
	close(c.out)
}
