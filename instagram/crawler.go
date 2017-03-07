package instagram

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/icholy/nick_bot/model"
	log "github.com/sirupsen/logrus"
)

type Crawler struct {
	username string
	password string

	users     []*model.User
	userIndex int

	out  chan *model.Media
	stop chan struct{}
}

func NewCrawler(username, password string) *Crawler {
	c := &Crawler{
		username: username,
		password: password,

		out:  make(chan *model.Media),
		stop: make(chan struct{}),
	}
	go c.loop()
	return c
}

func (c *Crawler) Media() <-chan *model.Media {
	return c.out
}

func (c *Crawler) loop() {
	for {
		if err := c.crawl(); err != nil {
			log.Printf("crawler: %s\n", err)
		}
		// sleep up to 5-20 minutes
		time.Sleep(
			5*time.Minute + time.Duration(rand.Intn(15))*time.Minute,
		)
	}
}

func (c *Crawler) crawl() error {
	s, err := NewSession(c.username, c.password)
	if err != nil {
		return err
	}
	defer s.Close()
	user, err := c.getNextUser(s)
	if err != nil {
		return err
	}
	medias, err := s.GetRecentUserMedias(user)
	if err != nil {
		return err
	}
	log.Printf("crawler: found %d media item(s) for %s\n", len(medias), user.Name)
	for _, media := range medias {
		c.out <- media
	}
	return nil
}

func (c *Crawler) getNextUser(s *Session) (*model.User, error) {

	// fetch the users again if we've used them all or there are none
	if c.userIndex >= len(c.users) {
		users, err := s.GetUsers()
		if err != nil {
			return nil, err
		}
		model.ShuffelUsers(users)
		c.users = users
	}

	if len(c.users) == 0 {
		return nil, fmt.Errorf("no users")
	}

	user := c.users[c.userIndex]
	c.userIndex++
	return user, nil
}
