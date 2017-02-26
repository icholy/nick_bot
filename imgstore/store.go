package imgstore

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/model"
)

type Store struct {
	db *sql.DB
	m  sync.Mutex
}

func Open(database string) (*Store, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.initDatabase(); err != nil {
		db.Close()
		return nil, err
	}
	return s, err
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initDatabase() error {
	s.m.Lock()
	defer s.m.Unlock()
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			media_id    TEXT,
			media_url   TEXT,
			user_id     TEXT,
			user_name   TEXT,
			like_count  INTEGER,
			face_count  INTEGER,
			posted_at   INTEGER,
			state       INTEGER
		)
	`)
	return err
}

func (s *Store) Put(rec *model.Record) error {
	s.m.Lock()
	defer s.m.Unlock()
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
	s.m.Lock()
	defer s.m.Unlock()
	row := s.db.QueryRow(
		`SELECT * FROM media WHERE media_id = ? LIMIT 1`, id,
	)
	return scanRecord(row)
}

func (s *Store) Has(id string) (bool, error) {
	s.m.Lock()
	defer s.m.Unlock()
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(1) FROM media WHERE media_id = ? LIMIT 1`, id,
	).Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (s *Store) SetState(id string, state model.MediaState) error {
	s.m.Lock()
	defer s.m.Unlock()
	resp, err := s.db.Exec(
		`UPDATE media SET state = ? WHERE media_id = ?`,
		state, id,
	)
	if err != nil {
		return err
	}
	n, err := resp.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("media not found: %s, %d", id)
	}
	return nil
}

func (s *Store) Stats(state model.MediaState) (Stats, error) {
	s.m.Lock()
	defer s.m.Unlock()
	rows, err := s.db.Query(`
		SELECT COUNT(1), face_count
		FROM media
		WHERE state = ?
		GROUP BY face_count
	`, state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats Stats
	for rows.Next() {
		var s Stat
		if err := rows.Scan(&s.Count, &s.Faces); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *Store) ResetStates() error {
	s.m.Lock()
	defer s.m.Unlock()
	_, err := s.db.Exec(
		`UPDATE media SET state = ?`,
		model.MediaAvailable,
	)
	return err
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
