package mmap

type MapFunc[T any, R any] func(t T) R

func Map[T any, R any](ts []T, mapfunc MapFunc[T, R]) []R {
	rs := make([]R, len(ts))
	for _, t := range ts {
		rs = append(rs, mapfunc(t))
	}
	return rs
}
