package imgstore

import (
	"math/rand"
)

type SearchStrategy int

const (
	FacesGlobalStrategy SearchStrategy = iota
	LikesGlobalStrategy
	ScoreGlobalStrategy
	FacesUserStrategy
	LikesUserStrategy
	ScoreUserStrategy
	RandomStrategy
)

func (s SearchStrategy) String() string {
	switch s {
	case FacesGlobalStrategy:
		return "FacesGlobal"
	case LikesGlobalStrategy:
		return "LikesGlobal"
	case ScoreGlobalStrategy:
		return "ScoreGlobal"
	case FacesUserStrategy:
		return "FacesUser"
	case LikesUserStrategy:
		return "LikesUser"
	case ScoreUserStrategy:
		return "ScoreUser"
	case RandomStrategy:
		return "Random"
	default:
		return "invalid"
	}
}

var strategies = []struct {
	P int
	S SearchStrategy
}{
	{10, FacesGlobalStrategy},
	{10, LikesGlobalStrategy},
	{10, ScoreGlobalStrategy},
	{20, FacesUserStrategy},
	{20, LikesUserStrategy},
	{20, ScoreUserStrategy},
	{5, RandomStrategy},
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
