package mslice

func Flatten[T any](arr [][]T) []T {
	total := 0
	for _, s := range arr {
		total += len(s)
	}

	result := make([]T, 0, total)
	for _, s := range arr {
		result = append(result, s...)
	}
	return result
}

func Filter[T any](ts []T, keep func(T) bool) []T {
	filteredTS := []T{}
	for _, t := range ts {
		if keep(t) {
			filteredTS = append(filteredTS, t)
		}
	}
	return filteredTS
}
