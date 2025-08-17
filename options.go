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
		apply(*registry) error
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
	optionFunc func(*registry) error

	// optionsPriority defines the priority of an option, which determines its order of execution
	optionPriority int
)

const (
	priorityUndefined optionPriority = 0

	priorityHighest = math.MaxInt - iota
	prioritySecondHighest
	priorityThirdHighest

	priorityLowest = math.MinInt + iota
	prioritySecondLowest
	priorityThirdLowest
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
func (t *optionsBuilder) apply(r *registry) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := applyOptions(r, t.o...); err != nil {
		return err
	}

	if t.isGlobalInstance {
		t.o = nil
	}

	return nil
}

// apply a single option
func (t *option) apply(r *registry) error {
	return t.optionFunc(r)
}

// applyOptions applies all the provided options to the registry/call
func applyOptions(r *registry, opts ...*option) (err error) {
	r.ensureCallOpts()

	if len(opts) == 0 {
		return
	}

	slices.SortStableFunc(opts, optionSorter)

	for _, opt := range opts {
		if err = opt.apply(r); err != nil {
			r.dropCallOpts()
			return err
		}

	}

	return
}

// unwrapOptions unwrap []Option interface to []*option from concrete []*optionsBuilder
func unwrapOptions(opts []Option) []*option {
	unwrapped := make([]*option, 0, len(opts))

	for _, opt := range opts {
		ob := opt.(*optionsBuilder)
		unwrapped = append(unwrapped, ob.o...)
	}

	return unwrapped
}

func optionSorter(a, b *option) int {
	if a.optionPriority > b.optionPriority {
		return -1
	}

	if a.optionPriority < b.optionPriority {
		return 1
	}

	return 0
}
