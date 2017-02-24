package instagram

import (
	"errors"
	"fmt"
	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/response"
	"time"
)

var ErrInvalidResponseStatus = errors.New("instagram: invalid response status")

type User struct {
	ID   string
	Name string
}

type Image struct {
	ID        string
	URL       string
	LikeCount int
	PostedAt  time.Time
}

func (m *Image) String() string {
	return fmt.Sprintf("Image: [%d likes] @%s %s",
		m.LikeCount, m.PostedAt, m.URL,
	)
}

type Session struct {
	insta *goinsta.Instagram
}

func New(username, password string) (*Session, error) {
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

func (Session) getLargestImage(info response.MediaInfoResponse) (*Image, error) {
	if len(info.Items) == 0 {
		return nil, errors.New("no items in media info")
	}
	var (
		item   = info.Items[0]
		images = item.ImageVersions2.Candidates
	)
	if len(images) == 0 {
		return nil, errors.New("no image candidates")
	}

	// find the largest image
	m := images[0]
	for _, v := range images {
		if v.Width*v.Height > m.Width*m.Height {
			m = v
		}
	}
	return &Image{
		ID:        item.ID,
		URL:       m.URL,
		LikeCount: item.LikeCount,
		PostedAt:  time.Unix(int64(item.Caption.CreatedAt), 0),
	}, nil
}

func (s *Session) GetRecentUserImages(u *User) ([]*Image, error) {
	resp, err := s.insta.FirstUserFeed(u.ID)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	var images []*Image
	for _, item := range resp.Items {
		candidates := item.ImageVersions2.Candidates
		if len(candidates) == 0 {
			continue
		}
		// choose the largest version of the image
		m := candidates[0]
		for _, c := range candidates {
			if c.Width*c.Height > m.Width*m.Height {
				m = c
			}
		}
		images = append(images, &Image{
			ID:        item.ID,
			URL:       m.URL,
			LikeCount: item.LikeCount,
			PostedAt:  time.Unix(int64(item.Caption.CreatedAt), 0),
		})
	}
	return images, nil
}

func (s *Session) GetUserMediaIDS(u *User) ([]string, error) {
	resp, err := s.insta.FirstUserFeed(u.ID)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	var ids []string
	for _, item := range resp.Items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (s *Session) GetUserImage(mediaID string) (*Image, error) {
	resp, err := s.insta.MediaInfo(mediaID)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	m, err := s.getLargestImage(resp)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Session) GetUsers() ([]*User, error) {
	id := s.insta.LoggedInUser.StringID()
	resp, err := s.insta.UserFollowing(id, "")
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, ErrInvalidResponseStatus
	}
	var users []*User
	for _, u := range resp.Users {
		users = append(users, &User{
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
