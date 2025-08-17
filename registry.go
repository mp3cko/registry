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

	"github.com/mp3cko/registry/access"
)

var (
	defaultRegistry, _ = NewRegistry()
)

// NewRegistry creates a new registry with the given options.
//
// Do not copy registries, instead use:
//
//	NewRegistry(WithCloneRegistry(src))
func NewRegistry(opts ...Option) (*registry, error) {
	reg := &registry{
		store: map[reflect.Type]map[string]any{},
		config: &registryConfig{
			accessibility: access.AccessibleInsidePackage,
			namedness:     access.NamednessUndefined,
		},
	}

	if err := applyOptions(reg, unwrapOptions(opts)...); err != nil {
		return nil, err
	}

	reg.config.initComplete = true

	return reg, nil
}

// SetDefaultRegistry changes the default registry used for all operations
func SetDefaultRegistry(r *registry) {
	old := defaultRegistry

	old.mu.Lock()
	defer old.mu.Unlock()

	defaultRegistry = r
}

// Set registers a instance inside the registry, modify its behavior by passing in options.
//
// Simplest Example:
//
//	Set(value)
//
// Less simple Example:
//
//	Set[InterfaceType](concreteValue, WithName("ConcreteInterface"))
//
// Complex Example:
//
//	Set(
//		myService,
//		WithRegistry(serviceRegistry).
//		WithUniqueName().
//		WithUniqueType().
//		WithName("ExternalService"),
//	)
func Set[T any](val T, opts ...Option) error {
	r := defaultRegistry
	r.mu.Lock()

	if err := applyOptions(r, unwrapOptions(opts)...); err != nil {
		r.mu.Unlock()
		return err
	}

	if r.callOptions.withRegistry != nil {
		withReg := r.callOptions.withRegistry
		callOpts := r.callOptions.withoutRegistry()

		r.dropCallOpts()
		r.mu.Unlock()

		r = withReg
		r.mu.Lock()
		r.callOptions = callOpts
	}

	defer r.mu.Unlock()

	return setType(r, val)
}

// Get retrieves the registered instance from a registry.
// If no options are provided it will return the default registered instance or ErrNotFound if it doesn't exist.
// Its behavior can be modified by passing in options (WithName, WithRegistry...)
func Get[T any](opts ...Option) (T, error) {
	r := defaultRegistry
	r.mu.Lock()

	if err := applyOptions(r, unwrapOptions(opts)...); err != nil {
		r.cleanup()
		return zeroValue[T](), err
	}

	if r.callOptions.uniqueName {
		r.cleanup()
		return zeroValue[T](), fmt.Errorf("Get WithUniqueNames: %w", ErrNotSupported)
	}

	if r.callOptions.withRegistry != nil {
		withReg := r.callOptions.withRegistry
		callOpts := r.callOptions.withoutRegistry()

		r.dropCallOpts()
		r.mu.Unlock()

		r = withReg
		r.mu.Lock()
		r.callOptions = callOpts
	}

	defer r.cleanup()

	return getType[T](r)
}

func GetAll(opts ...Option) (map[reflect.Type]map[string]any, error) {
	r := defaultRegistry
	r.mu.Lock()

	if err := applyOptions(r, unwrapOptions(opts)...); err != nil {
		r.cleanup()
		return nil, err
	}

	if r.callOptions.uniqueName {
		r.cleanup()
		return nil, fmt.Errorf("GetAll WithUniqueName: %w, use WithUniqueType instead", ErrNotSupported)
	}

	if r.callOptions.withRegistry != nil {
		withReg := r.callOptions.withRegistry
		callOpts := r.callOptions.withoutRegistry()

		r.dropCallOpts()
		r.mu.Unlock()

		r = withReg
		r.mu.Lock()
		r.callOptions = callOpts
	}

	defer r.cleanup()

	return getAll(r), nil
}

func Unset[T any](val T, opts ...Option) error {
	r := defaultRegistry
	r.mu.Lock()

	if err := applyOptions(r, unwrapOptions(opts)...); err != nil {
		r.cleanup()
		return err
	}

	if r.callOptions.uniqueName {
		r.cleanup()
		return fmt.Errorf("Unset WithUniqueName: %w, use WithUniqueType instead", ErrNotSupported)
	}

	if r.callOptions.withRegistry != nil {
		withReg := r.callOptions.withRegistry
		callOpts := r.callOptions.withoutRegistry()

		r.cleanup()

		r = withReg
		r.mu.Lock()
		r.callOptions = callOpts
	}

	defer r.cleanup()

	return unsetType(r, val)
}

// func UnsetAll

// setType in registry, caller must handle mutex locking
func setType[T any](r *registry, val T) error {
	if r.callOptions == nil {
		r.ensureCallOpts()
	}
	// take snapshots to avoid surprises if callOptions gets dropped later
	co := r.callOptions
	cfg := r.config

	name := valueOrDefault(co.name, cfg.defaultName)

	typeMustBeUnique := cfg.uniqueTypes || co.uniqueType
	nameMustBeUnique := cfg.uniqueNames || co.uniqueName

	rt := reflect.TypeFor[T]()

	typeNamedness, typeAccessibility := access.Info(zeroValue[T]())

	requiredAccessibility := max(cfg.accessibility, co.accessibility)
	if typeAccessibility < requiredAccessibility {
		return fmt.Errorf("Set '%T' failed: %w. Wanted at least '%s' but got '%s'", val, ErrAccessibilityTooLow, requiredAccessibility, typeAccessibility)
	}

	requiredNamedness := max(cfg.namedness, co.namedness)
	if typeNamedness < requiredNamedness {
		return fmt.Errorf("Set '%T' failed: %w. Wanted at least '%s' but got '%s'", val, ErrNamednessTooLow, requiredNamedness, typeNamedness)
	}

	if _, ok := r.store[rt]; !ok {
		r.store[rt] = map[string]any{}
	}

	if typeMustBeUnique && len(r.store[rt]) != 0 {
		if name != "" {
			return fmt.Errorf("Set '%T' named '%s' failed: %w", val, name, ErrNotUniqueType)
		}

		return fmt.Errorf("Set '%T' failed: %w", val, ErrNotUniqueType)
	}

	if _, ok := r.store[rt][name]; ok {
		if nameMustBeUnique {
			if name != "" {
				return fmt.Errorf("Set '%T' named '%s' failed: %w", val, name, ErrNotUniqueName)
			}

			return fmt.Errorf("Set '%T' failed: %w", val, ErrNotUniqueName)
		}
	}

	r.store[rt][name] = val

	return nil
}

// unsetType from the registry, caller must handle mutex locking
func unsetType[T any](r *registry, val T) error {
	r.ensureCallOpts()

	co := r.callOptions
	cfg := r.config

	typeMustBeUnique := co.uniqueType

	name := valueOrDefault(co.name, cfg.defaultName)
	rt := reflect.TypeFor[T]()

	instances, ok := r.store[rt]
	if !ok {
		return fmt.Errorf("Unset '%T' failed: %w", val, ErrNotFound)
	}

	if typeMustBeUnique && len(instances) > 1 {
		return fmt.Errorf("Unset '%T' WithUniqueType failed: %w", val, ErrNotUniqueType)
	}

	if _, ok := instances[name]; !ok {
		if name != cfg.defaultName {
			return fmt.Errorf("Unset '%T' named '%s' failed: %w", val, name, ErrNotFound)
		}

		return fmt.Errorf("Unset '%T' failed: %w", val, ErrNotFound)
	}

	if len(instances) > 1 {
		delete(instances, name)
	} else {
		delete(r.store, rt)
	}

	return nil
}

// getType from the registry, caller must handle mutex locking
func getType[T any](r *registry) (T, error) {
	r.ensureCallOpts()

	co := r.callOptions
	cfg := r.config

	if co.uniqueName {
		return zeroValue[T](), fmt.Errorf("Get '%T' failed: %w", zeroValue[T](), ErrNotSupported)
	}

	typeMustBeUnique := cfg.uniqueTypes || co.uniqueType

	name := valueOrDefault(co.name, cfg.defaultName)
	rt := reflect.TypeFor[T]()

	instances, ok := r.store[rt]
	if !ok {
		z := zeroValue[T]()

		return z, fmt.Errorf("Get '%T' failed: %w", z, ErrNotFound)
	}

	if typeMustBeUnique && len(instances) > 1 {
		z := zeroValue[T]()
		if name != "" {
			return z, fmt.Errorf("Get '%T' named '%s' failed: %w", z, name, ErrNotUniqueType)
		}

		return z, fmt.Errorf("Get '%T' failed: %w", zeroValue[T](), ErrNotUniqueType)
	}

	val, ok := r.store[rt][name]
	if !ok {
		z := zeroValue[T]()
		if name == "" {
			return z, fmt.Errorf("Get '%T' failed: %w", z, ErrNotFound)
		}

		return z, fmt.Errorf("Get '%T' named '%s' failed: %w", z, name, ErrNotFound)
	}

	return val.(T), nil
}

// getAll returns all registered instances, filtered by callopts from r
func getAll(r *registry) map[reflect.Type]map[string]any {
	defer r.dropCallOpts()

	stub := &registry{
		config: r.config.clone(),
	}

	cloneEntries(r, stub)

	return stub.store
}
