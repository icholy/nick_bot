package model

import "math/rand"

func ShuffelUsers(users []*User) {
	for i := range users {
		j := rand.Intn(i + 1)
		users[i], users[j] = users[j], users[i]
	}
}
