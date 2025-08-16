package reg

import (
	"fmt"
	"reflect"
)

// WithRegistry implementation
func withRegistryOption(r *registry) *option {
	f := func(original *registry) (*registry, error) {
		if !original.opts.initDone {
			return nil, fmt.Errorf("WithRegistry used inside NewRegistry: %w", ErrNotSupported)
		}
		return r, nil
	}

	return newOptionWithPriority(f, priorityHighest)
}

// WithUniqueInstance implementation
func withUniqueInstanceOption() *option {
	f := func(r *registry) (*registry, error) {
		if !r.opts.initDone {
			r.opts.uniqueInstanceCall = true
		} else {
			r.opts.uniqueInstance = true
		}

		return r, nil
	}

	return newOption(f)
}

// WithUniqueName implementation
func withUniqueNameOption() *option {
	f := func(r *registry) (*registry, error) {
		if !r.opts.initDone {
			r.opts.uniqueName = true
		} else {
			r.opts.uniqueNameCall = true
		}

		return r, nil
	}

	return newOption(f)
}

// WithName implementation
func withNameOption(n string) *option {
	f := func(r *registry) (*registry, error) {
		if !r.opts.initDone {
			r.opts.nameDefault = n
		} else {
			r.opts.nameCall = n
		}

		return r, nil
	}

	return newOption(f)
}

// WithCloneStore implementation
func withCloneStoreOption(src *registry) *option {
	f := func(r *registry) (*registry, error) {
		if r.opts.initDone {
			return nil, fmt.Errorf("WithCloneStore used outside NewRegistry: %w", ErrNotSupported)
		}

		if src == nil {
			return r, nil
		}

		// Deep copy store from src to r
		src.mu.RLock()
		defer src.mu.RUnlock()

		if r.store == nil {
			r.store = make(map[reflect.Type]map[string]any, len(src.store))
		}

		for types, vals := range src.store {
			typeMap, ok := r.store[types]
			if !ok {
				typeMap = make(map[string]any, len(vals))
				r.store[types] = typeMap
			}

			for name, val := range vals {
				if r.opts.nameDefault != src.opts.nameDefault && name == src.opts.nameDefault {
					typeMap[r.opts.nameDefault] = val
				} else {
					typeMap[name] = val
				}
			}
		}

		return r, nil
	}

	return newOptionWithPriority(f, priorityLowest)
}

func newOption(o optionFunc) *option {
	return &option{
		optionFunc: o,
	}
}

func newOptionWithPriority(o optionFunc, p optionPriority) *option {
	return &option{
		optionFunc:     o,
		optionPriority: p,
	}
}
