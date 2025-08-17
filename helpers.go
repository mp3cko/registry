package reg

// valueOrDefault returns default if val == zeroValue[T]()
func valueOrDefault[T comparable](val, def T) T {
	if val == zeroValue[T]() {
		return def
	}

	return val
}

// zeroValue returns the zero value of type T
func zeroValue[T any]() T {
	var zero T
	return zero
}
