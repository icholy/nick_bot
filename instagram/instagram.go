package instagram

import (
	"errors"
	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/response"
)

var ErrInvalidResponseStatus = errors.New("instagram: invalid response status")

type User struct {
	ID   string
	Name string
}

type Image struct {
	ID  string
	URL string
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
		ID:  item.ID,
		URL: m.URL,
	}, nil
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
