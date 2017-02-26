package model

import (
	"fmt"
	"time"
)

type User struct {
	ID   string
	Name string
}

type Media struct {
	ID        string
	URL       string
	UserID    string
	Username  string
	LikeCount int
	PostedAt  time.Time
}

func (m *Media) String() string {
	return fmt.Sprintf("Media: [%d likes] @%s %s",
		m.LikeCount, m.Username, m.URL,
	)
}

type MediaState int

const (
	MediaAvailable MediaState = iota
	MediaRejected
	MediaUsed
)

type Record struct {
	Media
	FaceCount int
	State     MediaState
}

func (rec *Record) String() string {
	return fmt.Sprintf("Record: [%d face(s)] [%d like(s)] @%s %s",
		rec.FaceCount,
		rec.LikeCount,
		rec.Username,
		rec.URL,
	)
}
