package sliceutil

func IfNotExist[T comparable](slice []T, val T) []T {
	for _, e := range slice {
		if e == val {
			return slice
		}
	}
	return append(slice, val)
}
