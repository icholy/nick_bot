package imgstore

import (
	"math/rand"
)

type SearchStrategy int

const (
	MostFacesGlobalStrategy SearchStrategy = iota
	MostLikesGlobalStrategy
	MostScoreGlobalStrategy
	MostFacesUserStrategy
	MostLikesUserStrategy
	MostScoreUserStrategy
)

func (s SearchStrategy) String() string {
	switch s {
	case MostFacesGlobalStrategy:
		return "MostFacesGlobal"
	case MostLikesGlobalStrategy:
		return "MostLikesGlobal"
	case MostScoreGlobalStrategy:
		return "MostScoreGlobal"
	case MostFacesUserStrategy:
		return "MostFacesUser"
	case MostLikesUserStrategy:
		return "MostLikesUser"
	case MostScoreUserStrategy:
		return "MostScoreUser"
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
	{10, MostScoreGlobalStrategy},
	{30, MostFacesUserStrategy},
	{20, MostLikesUserStrategy},
	{20, MostScoreUserStrategy},
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
