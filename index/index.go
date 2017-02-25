package index

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/model"
)

type Index struct {
	db *sql.DB
}

func Open(database string) (*Index, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, err
	}
	return &Index{db}, nil
}

func (mi *Index) Close() error {
	return mi.db.Close()
}

func (mi *Index) CreateDatabase() error {
	_, err := mi.db.Exec(`
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

func (mi *Index) Put(rec *model.Record) error {
	_, err := mi.db.Exec(
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

func (mi *Index) Get(id string) (*model.Record, error) {
	row := mi.db.QueryRow(
		`SELECT * FROM media WHERE media_id = ? LIMIT 1`, id,
	)
	return scanRecord(row)
}

func (mi *Index) Has(id string) (bool, error) {
	var count int
	if err := mi.db.QueryRow(
		`SELECT COUNT(1) FROM media WHERE media_id = ? LIMIT 1`, id,
	).Scan(&count); err != nil {
		return false, err
	}
	return count == 1, nil
}

func (mi *Index) SetState(id string, state model.MediaState) error {
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

func (mi *Index) Search(minFaces int) (*model.Record, error) {
	row := mi.db.QueryRow(`
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
