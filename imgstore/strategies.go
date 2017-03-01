package imgstore

import (
	"math/rand"
)

type SearchStrategy int

const (
	TopFacesStrategy SearchStrategy = iota
	TopLikesStrategy
	FacesUserStrategy
	LikesUserStrategy
)

func (s SearchStrategy) String() string {
	switch s {
	case TopFacesStrategy:
		return "TopFaces"
	case TopLikesStrategy:
		return "TopLikes"
	case FacesUserStrategy:
		return "FacesUser"
	case LikesUserStrategy:
		return "LikesUser"
	default:
		return "invalid"
	}
}

var strategies = []struct {
	P int
	S SearchStrategy
}{
	{40, TopFacesStrategy},
	{40, TopLikesStrategy},
	{10, FacesUserStrategy},
	{10, LikesUserStrategy},
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

func ChooseStrategy() SearchStrategy {
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
