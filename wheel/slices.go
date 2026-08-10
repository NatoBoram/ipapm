package wheel

func Find[T any](v []T, f func(T) bool) (T, bool) {
	for _, v := range v {
		if ok := f(v); ok {
			return v, true
		}
	}

	var x T
	return x, false
}
