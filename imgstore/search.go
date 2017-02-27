package imgstore

import (
	"errors"
	"log"

	"github.com/icholy/nick_bot/model"
)

func (s *Store) SearchRandom(minFaces int) (*model.Record, error) {
	strategy := ChooseStrategy()
	log.Printf("imgstore: using %s strategy\n", strategy)
	return s.Search(minFaces, strategy)
}

func (s *Store) Search(minFaces int, strategy SearchStrategy) (*model.Record, error) {
	switch strategy {
	case MostFacesGlobalStrategy:
		return s.searchMostFacesGlobal(minFaces)
	case MostLikesGlobalStrategy:
		return s.searchMostLikesGlobal(minFaces)
	case MostLikesUserStrategy:
		return s.searchMostLikesUser(minFaces)
	case MostFacesUserStrategy:
		return s.searchMostFacesUser(minFaces)
	default:
		return nil, errors.New("strategy not implemented")
	}
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