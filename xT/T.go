package xT

// Contains checks if str is in list.
func Contains[T comparable](slices []T, t T) bool {
	for _, each := range slices {
		if each == t {
			return true
		}
	}
	return false
}

// Remove removes given strs from strings.
func Remove[T comparable](slices []T, ts ...T) []T {
	out := append([]T(nil), slices...)
	for _, t := range ts {
		var n int
		for _, v := range out {
			if v != t {
				out[n] = v
				n++
			}
		}
		out = out[:n]
	}
	return out
}
