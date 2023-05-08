package lib

func isPresent[T comparable](element T, set []T) bool {
	for _, e := range set {
		if e == element {
			return true
		}
	}
	return false
}
func Contains[T comparable](set1 []T, set2 []T) bool {
	for i := range set2 {
		if !isPresent(set2[i], set1) {
			return false
		}
	}
	return true
}
