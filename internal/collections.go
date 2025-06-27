package internal

func Map[T, U any](source []T, f func(T) U) []U {
	result := make([]U, len(source))
	for i, item := range source {
		result[i] = f(item)
	}
	return result
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

// Curried versions

func CurriedMap[T, U any](f func(T) U) func([]T) []U {
	return func(source []T) []U {
		return Map(source, f)
	}
}

func CurriedFind[T any](predicate func(T) bool) func([]T) (*T, bool) {
	return func(slice []T) (*T, bool) {
		return Find(slice, predicate)
	}
}

func CurriedFilter[T any](predicate func(T) bool) func([]T) []T {
	return func(slice []T) []T {
		return Filter(slice, predicate)
	}
}

// MapWithError applies a function that can return an error to each element.
func MapWithError[T, U any](source []T, f func(T) (U, error)) ([]U, error) {
	result := make([]U, len(source))
	for i, item := range source {
		mappedItem, err := f(item)
		if err != nil {
			return nil, err
		}
		result[i] = mappedItem
	}
	return result, nil
}

// Partition splits a slice into chunks of a specified size.
func Partition[T any](slice []T, size int) [][]T {
	if size <= 0 {
		panic("partition size must be positive")
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}
