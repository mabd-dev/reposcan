package mmap

type MapFunc[T any, R any] func(t T) R

func Map[T any, R any](ts []T, mapFunc MapFunc[T, R]) []R {
	rs := make([]R, len(ts))
	for i, t := range ts {
		rs[i] = mapFunc(t)
	}
	return rs
}
