package manager

import "math/rand"

func filter[T comparable](arr []T, predicate func(T) bool) []T {
	newArr := make([]T, 0)
	for _, elm := range arr {
		if predicate(elm) {
			newArr = append(newArr, elm)
		}
	}

	return newArr
}

func randomName() string {
	name := ""
	for i := 0; i < 10; i++ {
		name += string(rand.Intn('z'-'a') + 'a')
	}
	return name
}
