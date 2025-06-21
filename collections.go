package main

func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i, t := range ts {
		us[i] = f(t)
	}
	return us
}

func Find[T any](slice []T, predicate func(T) bool) (*T, bool) {
	for i := range slice {
		if predicate(slice[i]) {
			return &slice[i], true
		}
	}
	return nil, false
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}
