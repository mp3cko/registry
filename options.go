package reg

import (
	"math"
	"slices"
	"sync"
)

type (
	// Option for configuring the registry.
	//
	// Some Options behave differently if used as part of the constructor or specific call and may not be supported in both.
	// Check the doc for each to understand how it works.
	//
	// All the provided options can be passed individually (variadic arguments) or chained one after another
	Option interface {
		apply(*registry) (*registry, error)
	}

	// optionsBuilder is a wrapper that allows chaining multiple option values directly instead of passing them individually, so both of these are valid:
	//
	//	reg.Set(val, WithName("example").WithRegistry(r))
	//	reg.Set(val, WithName("example"), WithRegistry(r))
	optionsBuilder struct {
		mu               sync.Mutex
		o                []*option
		isGlobalInstance bool
	}

	// single option representation
	option struct {
		optionFunc
		optionPriority
	}

	// optionFunc is an implementation of Option
	optionFunc func(*registry) (*registry, error)

	// optionsPriority defines the priority of an option, which determines its order of execution
	optionPriority int
)

const (
	priorityUndefined optionPriority = iota
	priorityHighest

	priorityLowest = math.MinInt
)

// @TODO - see if I want to use this
// var (
// 	 // Global options instance, it resets its Option chain after each apply
// 	Options = &options{isGlobalInstance: true}
// )

func (t *optionsBuilder) and(opts ...*option) *optionsBuilder {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.o = append(t.o, opts...)

	return t
}

// apply applies all contained options in the following order:
// 1. Their priority
// 2. Their order of appearance
func (t *optionsBuilder) apply(r *registry) (*registry, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r, err := applyOptions(r, t.o...)
	if err != nil {
		return nil, err
	}

	if t.isGlobalInstance {
		t.o = nil
	}

	return r, nil
}

// apply a single option
func (t *option) apply(r *registry) (*registry, error) {
	return t.optionFunc(r)
}

// applyOptions applies all the provided options to the registry/call
func applyOptions(reg *registry, opts ...*option) (*registry, error) {
	if len(opts) == 0 {
		return reg, nil
	}

	slices.SortStableFunc(opts, func(a, b *option) int {
		if a.optionPriority > b.optionPriority {
			return -1
		} else if a.optionPriority < b.optionPriority {
			return 1
		}

		return 0
	})

	for _, opt := range opts {
		newReg, err := opt.apply(reg)
		if err != nil {
			return nil, err
		}

		reg = newReg
	}

	return reg, nil
}

// unwrapOptions unwrap []Option interface to []*option from concrete []*optionsBuilder
func unwrapOptions(opts []Option) []*option {
	var unwrapped []*option

	for _, opt := range opts {
		ob := opt.(*optionsBuilder)
		unwrapped = append(unwrapped, ob.o...)
	}

	return unwrapped
}
