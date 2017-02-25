package imgstore

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/model"
)

type Store struct {
	db *sql.DB
}

func Open(database string) (*Store, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}
	return &Store{db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) CreateDatabase() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			media_id    TEXT,
			media_url,  TEXT,
			user_id     TEXT,
			user_name   TEXT,
			like_count  INTEGER,
			face_count  INTEGER,
			posted_at  INTEGER,
			media_state INTEGER
		)
	`)
	return err
}

func (s *Store) Put(rec *model.Record) error {
	_, err := s.db.Exec(
		`INSERT INTO media VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID,
		rec.URL,
		rec.UserID,
		rec.Username,
		rec.LikeCount,
		rec.FaceCount,
		rec.PostedAt.Unix(),
		rec.State,
	)
	return err
}

func (s *Store) Get(id string) (*model.Record, error) {
	row := s.db.QueryRow(
		`SELECT * FROM media WHERE media_id = ? LIMIT 1`, id,
	)
	return scanRecord(row)
}

func (s *Store) Has(id string) (bool, error) {
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(1) FROM media WHERE media_id = ? LIMIT 1`, id,
	).Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (s *Store) SetState(id string, state model.MediaState) error {
	resp, err := s.db.Exec(
		`UPDATE media SET state = ? WHERE media_id = ? LIMIT 1`,
		state, id,
	)
	if err != nil {
		return err
	}
	n, err := resp.RowsAffected()
	if err != nil {
		return err
	}
	if n == 1 {
		return fmt.Errorf("media not found: %s", id)
	}
	return nil
}

func (s *Store) Search(minFaces int) (*model.Record, error) {
	row := s.db.QueryRow(`
		SELECT *
		FROM media
		WHERE state == ? AND face_count >= ?
		ORDER BY face_count ASC, like_count ASC
		LIMIT 1
	`, model.MediaAvailable, minFaces)
	return scanRecord(row)
}

func scanRecord(row *sql.Row) (*model.Record, error) {
	var (
		rec      model.Record
		postedAt int64
	)
	if err := row.Scan(
		&rec.ID,
		&rec.URL,
		&rec.UserID,
		&rec.Username,
		&rec.LikeCount,
		&rec.FaceCount,
		&postedAt,
		&rec.State,
	); err != nil {
		return nil, err
	}
	rec.PostedAt = time.Unix(postedAt, 0)
	return &rec, nil
}
