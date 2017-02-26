package imgstore

import (
	"errors"
	"math/rand"

	"github.com/icholy/nick_bot/model"
)

type SearchStrategy int

const (
	MostFacesGlobalStrategy SearchStrategy = iota
	MostLikesGlobalStrategy
	MostFacesUserStrategy
	MostLikesUserStrategy
)

var strategies = []struct {
	P int
	S SearchStrategy
}{
	{10, MostFacesGlobalStrategy},
	{10, MostLikesGlobalStrategy},
	{40, MostFacesUserStrategy},
	{40, MostLikesUserStrategy},
}

func init() {
	var total int
	for _, s := range strategies {
		total += s.P
	}
	if total != 100 {
		panic("strategy probabilities don't total 100")
	}
}

func (s *Store) selectStrategy() SearchStrategy {
	var (
		prob  = rand.Intn(100)
		total int
	)
	for _, strategy := range strategies {
		total += strategy.P
		if prob <= total {
			return strategy.S
		}
	}
	panic("should never happen")
}

func (s *Store) Search(minFaces int, strategy SearchStrategy) (*model.Record, error) {
	switch strategy {
	case MostFacesGlobalStrategy:
		return s.searchMostFacesGlobal(minFaces)
	case MostLikesGlobalStrategy:
		return s.searchMostLikesGlobal(minFaces)
	case MostLikesUserStrategy:
		return s.searchMostLikesUser(minFaces)
	default:
		return nil, errors.New("strategy not implemented")
	}
}

func (s *Store) SearchRandom(minFaces int) (*model.Record, error) {
	strategy := s.selectStrategy()
	return s.Search(minFaces, strategy)
}

func (s *Store) searchMostFacesGlobal(minFaces int) (*model.Record, error) {
	s.m.Lock()
	defer s.m.Unlock()
	row := s.db.QueryRow(`
		SELECT *
		FROM media
		WHERE state = ? AND face_count >= ?
		ORDER BY face_count DESC, like_count DESC
		LIMIT 1
	`, model.MediaAvailable, minFaces)
	return scanRecord(row)
}

func (s *Store) searchMostLikesGlobal(minFaces int) (*model.Record, error) {
	s.m.Lock()
	defer s.m.Unlock()
	row := s.db.QueryRow(`
		SELECT *
		FROM media
		WHERE state = ? AND face_count >= ?
		ORDER BY like_count DESC, face_count DESC
		LIMIT 1
	`, model.MediaAvailable, minFaces)
	return scanRecord(row)
}

func (s *Store) searchMostLikesUser(minFaces int) (*model.Record, error) {
	user, err := s.randomUserWithPhotos(minFaces)
	if err != nil {
		return nil, err
	}
	s.m.Lock()
	defer s.m.Unlock()
	row := s.db.QueryRow(`
		SELECT *
		FROM media
		WHERE state = ? AND face_count >= ? AND user_id = ?
		ORDER BY face_count DESC, face_count DESC
		LIMIT 1
	`, model.MediaAvailable, minFaces, user.ID)
	return scanRecord(row)
}

func (s *Store) searchMostFacesUser(minFaces int) (*model.Record, error) {
	user, err := s.randomUserWithPhotos(minFaces)
	if err != nil {
		return nil, err
	}
	s.m.Lock()
	defer s.m.Unlock()
	row := s.db.QueryRow(`
		SELECT *
		FROM media
		WHERE state = ? AND face_count >= ? AND user_id = ?
		ORDER BY like_count DESC, face_count DESC
		LIMIT 1
	`, model.MediaAvailable, minFaces, user.ID)
	return scanRecord(row)
}

func (s *Store) randomUserWithPhotos(minFaces int) (*model.User, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var u model.User
	if err := s.db.QueryRow(`
		SELECT user_id, user_name
		FROM media
		WHERE state = ? AND face_count >= ?
		GROUP BY user_id, user_name
		ORDER BY RANDOM()
		LIMIT 1
	`, model.MediaAvailable, minFaces,
	).Scan(&u.ID, &u.Name); err != nil {
		return nil, err
	}
	return &u, nil
}
