package imgstore

import (
	"math/rand"
)

type SearchStrategy int

const (
	MostFacesGlobalStrategy SearchStrategy = iota
	MostLikesGlobalStrategy
	MostFacesUserStrategy
	MostLikesUserStrategy
)

func (s SearchStrategy) String() string {
	switch s {
	case MostFacesGlobalStrategy:
		return "MostFacesGlobal"
	case MostLikesGlobalStrategy:
		return "MostLikesGlobal"
	case MostFacesUserStrategy:
		return "MostFacesUser"
	case MostLikesUserStrategy:
		return "MostLikesUser"
	default:
		return "invalid"
	}
}

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
