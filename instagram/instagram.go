package instagram

import (
	"errors"
	"net/url"
	"time"

	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/response"

	"github.com/icholy/nick_bot/model"
)

var ErrInvalidResponseStatus = errors.New("instagram: invalid response status")

type Session struct {
	insta *goinsta.Instagram
}

func NewSession(username, password string) (*Session, error) {
	insta := goinsta.New(username, password)
	if err := insta.Login(); err != nil {
		return nil, err
	}
	return &Session{
		insta: insta,
	}, nil
}

func (s *Session) Close() error {
	return s.insta.Logout()
}

func (Session) getLargestCandidate(candidates []response.ImageCandidate) response.ImageCandidate {
	m := candidates[0]
	for _, c := range candidates {
		if c.Width*c.Height > m.Width*m.Height {
			m = c
		}
	}
	return m
}

func (s *Session) GetRecentUserMedias(u *model.User) ([]*model.Media, error) {
	resp, err := s.insta.FirstUserFeed(u.ID)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	var images []*model.Media
	for _, item := range resp.Items {

		candidates := item.ImageVersions2.Candidates
		if len(candidates) == 0 {
			continue
		}
		m := s.getLargestCandidate(item.ImageVersions2.Candidates)

		// remove token from url
		mediaURL, err := s.cleanURL(m.URL)
		if err != nil {
			return nil, err
		}

		images = append(images, &model.Media{
			ID:        item.ID,
			URL:       mediaURL,
			UserID:    u.ID,
			Username:  u.Name,
			LikeCount: item.LikeCount,
			PostedAt:  time.Unix(int64(item.Caption.CreatedAt), 0),
		})
	}
	return images, nil
}

func (s *Session) GetUsers() ([]*model.User, error) {
	id := s.insta.LoggedInUser.StringID()
	resp, err := s.insta.UserFollowing(id, "")
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	var users []*model.User
	for _, u := range resp.Users {
		users = append(users, &model.User{
			ID:   u.StringID(),
			Name: u.Username,
		})
	}
	return users, nil
}

func (s *Session) UploadPhoto(imgPath string, caption string) error {
	resp, err := s.insta.UploadPhoto(imgPath, caption, s.insta.NewUploadID(), 100, 0)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (Session) cleanURL(rawurl string) (string, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	return u.String(), nil
}
