package imgstore

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"

	"github.com/icholy/nick_bot/model"
)

type LegacyMedia struct {
	ID        string
	URL       string
	UserID    string
	Username  string
	FaceCount int
}

func scanLegacyRow(rows *sql.Rows) (*LegacyMedia, error) {
	var media LegacyMedia
	if err := rows.Scan(
		&media.UserID,
		&media.Username,
		&media.ID,
		&media.URL,
		&media.FaceCount,
	); err != nil {
		return nil, err
	}
	return &media, nil
}

func convertLegacyToRecord(media *LegacyMedia) *model.Record {
	return &model.Record{
		Media: model.Media{
			ID:        media.ID,
			URL:       media.URL,
			Username:  media.Username,
			UserID:    media.UserID,
			LikeCount: 0,
		},
		FaceCount: media.FaceCount,
		State:     model.MediaUsed,
	}
}

func (s *Store) ImportLegacyDatabase(dbfile string) error {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT * FROM media`)
	if err != nil {
		return err
	}
	var count int64
	for rows.Next() {
		media, err := scanLegacyRow(rows)
		if err != nil {
			return err
		}
		rec := convertLegacyToRecord(media)
		if err := s.Put(rec); err != nil {
			return err
		}
		count++
	}
	log.Printf("Imported %d records\n", count)
	return rows.Err()
}
