package mmap

type MapFunc[T any, R any] func(t T) R

func Map[T any, R any](ts []T, mapFunc MapFunc[T, R]) []R {
	rs := make([]R, len(ts))
	for i, t := range ts {
		rs[i] = mapFunc(t)
	}
	return rs
}

// Keys returns slice of keys from a generic map
func Keys[T comparable, R any](m map[T]R) []T {
	keys := []T{}
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
