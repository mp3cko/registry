package reg

import "reflect"

// MustSet is a Set() helper that panics on error
func MustSet[T any](val T, opts ...Option) {
	if err := Set(val, opts...); err != nil {
		panic(err)
	}
}

// MustGet is a Get() helper that panics on error
func MustGet[T any](opts ...Option) T {
	out, err := Get[T](opts...)
	if err != nil {
		panic(err)
	}

	return out
}

// MustGetAll is a GetAll() helper that panics on error
func MustGetAll(opts ...Option) map[reflect.Type]map[string]any {
	out, err := GetAll(opts...)
	if err != nil {
		panic(err)
	}

	return out
}

// MustUnset is a Unset() helper that panics on error
func MustUnset[T any](val T, opts ...Option) {
	if err := Unset(val, opts...); err != nil {
		panic(err)
	}
}
