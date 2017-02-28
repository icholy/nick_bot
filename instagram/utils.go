package instagram

import (
	"github.com/icholy/nick_bot/model"
	"math/rand"
)

func shuffelUsers(users []*model.User) {
	for i := range users {
		j := rand.Intn(i + 1)
		users[i], users[j] = users[j], users[i]
	}
}
