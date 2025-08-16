// Package reg(istry) is an implementation of a service (or any type) registry using generics and reflection.
//
// You add and retrieve entries by their type and optionally by name.
//
// A basic example:
//
//	var t myType
//	err := reg.Set(t)
//	if err != nil {
//		// handle error
//	}
//
//	value, err := reg.Get[myType]()
//	if err != nil {
//		// handle error
//	}
//
// All the provided options can be passed in as variadic arguments or by chaining them directly. Ex:
//
//	reg.Get[myType](WithUnique().WithName("example"))
//
// or
//
//	reg.Get[myType](WithUnique(), WithName("example"))
//
// It is aliased to
//
//	reg
//
// for convenience.
package reg

import (
	"fmt"
	"reflect"
	"sync"
)

// To check for errors use errors.Is(err), don't use direct comparison(==) as they are wrapped
var (
	ErrNotFound  = fmt.Errorf("not found")
	ErrNotUnique = fmt.Errorf("not unique")

	ErrNotSupported = fmt.Errorf("not supported")
)

var defReg, _ = NewRegistry()

// NewRegistry creates a new registry with the given options.
//
// Registry instances must not be copied because they contains a mutex.
//
// Instead use:
//
//	NewRegistry(WithCloneStore(srcRegistry))
func NewRegistry(opts ...Option) (*registry, error) {
	reg := &registry{
		store: map[reflect.Type]map[string]any{},
		opts:  new(regOpts),
	}

	reg, err := applyOptions(reg, unwrapOptions(opts)...)
	if err != nil {
		return nil, err
	}

	reg.opts.initDone = true

	return reg, nil
}

// Set registers a typed instance inside the registry.
func Set[T any](val T, opts ...Option) error {
	var (
		err error
		r   = defReg
	)

	r, err = applyOptions(r, unwrapOptions(opts)...)
	if err != nil {
		return err
	}

	return setType(r, val)
}

// MustSet is a Set() wrapper that will panic on error
func MustSet[T any](val T, opts ...Option) {
	if err := Set(val, opts...); err != nil {
		panic(err)
	}
}

// Get retrieves the registered instance from a registry.
// If no options are provided it will return the default registered instance or ErrNotFound if it doesn't exist.
// Its behavior can be modified by passing in options (WithName, WithRegistry...)
func Get[T any](opts ...Option) (T, error) {
	var (
		err error
		r   = defReg
	)

	r, err = applyOptions(r, unwrapOptions(opts)...)
	if err != nil {
		return zeroType[T](), err
	}

	return getType[T](r)
}

// MustGet is a Get() wrapper that will panic on error
func MustGet[T any](opts ...Option) T {
	val, err := Get[T](opts...)
	if err != nil {
		panic(err)
	}

	return val
}

// registry is a type-safe registry where instances are registered and retrieved by type
// (and optionally by name if you want to register multiple instances of the same type).
type registry struct {
	mu sync.RWMutex
	// store maps a type to a map[name]instance. Default name is an empty string.
	store map[reflect.Type]map[string]any
	opts  *regOpts
}

type regOpts struct {
	initDone           bool   // indicates if the registry has been initialized
	nameDefault        string // registry-level default name
	nameCall           string // per-call name
	uniqueInstance     bool   // per-call unique instance constraint
	uniqueInstanceCall bool   // registry-level unique instance constraint
	uniqueName         bool   // per-call unique name constraint
	uniqueNameCall     bool   // registry-level unique name constraint

}

func (t *regOpts) resetCallFlags() {
	t.nameCall = ""
	t.uniqueInstanceCall = false
	t.uniqueNameCall = false
}

func setType[T any](r *registry, val T) error {
	r.mu.Lock()

	defer func() {
		r.opts.resetCallFlags()
		r.mu.Unlock()
	}()

	name := r.opts.nameDefault
	if r.opts.nameCall != "" {
		name = r.opts.nameCall
	}

	rt := reflect.TypeFor[T]()

	if _, ok := r.store[rt]; !ok {
		r.store[rt] = map[string]any{}
	}

	if r.opts.uniqueInstanceCall || r.opts.uniqueInstance {
		if _, ok := r.store[rt][name]; ok {
			if name == "" {
				return fmt.Errorf("register '%T' failed: %w", val, ErrNotUnique)
			}

			return fmt.Errorf("register '%T' name '%s' failed: %w", val, name, ErrNotUnique)
		}
	}

	r.store[rt][name] = any(val)

	return nil
}

func getType[T any](r *registry) (T, error) {
	r.mu.RLock()

	defer func() {
		r.opts.resetCallFlags()
		r.mu.RUnlock()
	}()

	name := r.opts.nameDefault
	if r.opts.nameCall != "" {
		name = r.opts.nameCall
	}

	rt := reflect.TypeFor[T]()

	typeMap, ok := r.store[rt]
	if !ok {
		z := zeroType[T]()
		return z, fmt.Errorf("get '%T' failed: %w", z, ErrNotFound)
	}

	svc, ok := typeMap[name]
	if !ok {
		z := zeroType[T]()
		return z, fmt.Errorf("get '%T' name '%s' failed: %w", z, name, ErrNotFound)
	}

	return svc.(T), nil
}

func zeroType[T any]() T {
	var zero T
	return zero
}
