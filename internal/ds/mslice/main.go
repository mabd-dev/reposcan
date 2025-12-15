package mslice

func Flatten[T any](arr [][]T) []T {
	flatArr := []T{}
	for _, inner := range arr {
		flatArr = append(flatArr, inner...)
	}
	return flatArr
}

type FilterFunc[T any] func(t T) bool

func Filter[T any](ts []T, filterFunc FilterFunc[T]) []T {
	filteredTS := []T{}
	for _, t := range ts {
		if filterFunc(t) {
			filteredTS = append(filteredTS, t)
		}
	}
	return filteredTS
}
