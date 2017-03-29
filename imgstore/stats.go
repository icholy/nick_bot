package imgstore

import (
	"fmt"
	"strings"
)

type Stat struct {
	Faces int   `json:"faces"`
	Count int64 `json:"count"`
}

func (s *Stat) String() string {
	return fmt.Sprintf("%d: %d face(s)", s.Count, s.Faces)
}

type Stats []Stat

func (s Stats) String() string {
	var ss []string
	for _, s := range s {
		ss = append(ss, s.String())
	}
	return strings.Join(ss, "\n")
}
