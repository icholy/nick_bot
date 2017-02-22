package main

import (
	"errors"
	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/response"
)

type User struct {
	ID   string
	Name string
}

type Image struct {
	ID  string
	URL string
}

type Stream struct {
	insta  *goinsta.Instagram
	Images chan string
}

func New(username, password string) (*Stream, error) {
	insta := goinsta.New(username, password)
	if err := insta.Login(); err != nil {
		return nil, err
	}
	return &Stream{
		insta:  insta,
		Images: make(chan string),
	}, nil
}

func (s *Stream) Close() error {
	return s.insta.Logout()
}

func (Stream) getLargestImage(info response.MediaInfoResponse) (*Image, error) {
	if len(info.Items) == 0 {
		return nil, errors.New("no items in media info")
	}
	item := info.Items[0]
	images := item.ImageVersions2.Candidates
	if len(images) == 0 {
		return nil, errors.New("no image condidates")
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

func (s *Stream) GetUserMediaIDS(u *User) ([]string, error) {
	resp, err := s.insta.FirstUserFeed(u.ID)
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, errors.New("invalid response code")
	}
	var ids []string
	for _, item := range resp.Items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

func (s *Stream) GetUserImage(mediaID string) (*Image, error) {
	media, err := s.insta.MediaInfo(mediaID)
	if err != nil {
		return nil, err
	}
	m, err := s.getLargestImage(media)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Stream) GetUsers() ([]*User, error) {
	id := s.insta.LoggedInUser.StringID()
	resp, err := s.insta.UserFollowing(id, "")
	if err != nil {
		return nil, err
	}
	if resp.Status != "ok" {
		return nil, errors.New("invalid response")
	}
	var users []*User
	for _, u := range resp.Users {
		users = append(users, &User{
			ID:   u.StringID(),
			Name: u.FullName,
		})
	}
	return users, nil
}

func (s *Stream) UploadPhoto(imgPath string, caption string) error {
	resp, err := s.insta.UploadPhoto(imgPath, caption, s.insta.NewUploadID(), 100, 0)
	if err != nil {
		return err
	}
	if resp.Status != "ok" {
		return errors.New("invalid status")
	}
	return nil
}

func main() {

}
