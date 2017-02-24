package mediaindex

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type Media struct {
	ID        string
	URL       string
	UserID    string
	Username  string
	LikeCount int
	FaceCount int
	CreateAt  time.Time
	State     MediaState
}

type MediaState int

const (
	MediaAvailable MediaState = iota
	MediaRejected
	MediaUsed
)

type MediaIndex struct {
	db *sql.DB
}

func Open(database string) (*MediaIndex, error) {
	db, err := sql.Open("sqlite3", "media.db")
	if err != nil {
		return nil, err
	}
	return &MediaIndex{db}, nil
}

func (mi *MediaIndex) Close() error {
	return mi.db.Close()
}

func (mi *MediaIndex) CreateDatabase() error {
	_, err := mi.db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			media_id    TEXT,
			media_url,  TEXT,
			user_id     TEXT,
			user_name   TEXT,
			face_count  INTEGER,
			created_at  INTEGER,
			media_state INTEGER
		)
	`)
	return err
}

func (mi *MediaIndex) Save(media *Media) error {
	_, err := mi.db.Exec(
		`INSERT INTO media VALUES (?, ?, ?, ?, ?, ?, ?)`,
		media.ID,
		media.URL,
		media.UserID,
		media.Username,
		media.FaceCount,
		media.CreateAt.Unix(),
		media.State,
	)
	return err
}

func (mi *MediaIndex) Has(id string) (bool, error) {
	var count int
	if err := mi.db.QueryRow(
		`SELECT COUNT(1) FROM media WHERE media_id = ? LIMIT 1`, id,
	).Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (mi *MediaIndex) Mark(id string, state MediaState) error {
	resp, err := mi.db.Exec(
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
