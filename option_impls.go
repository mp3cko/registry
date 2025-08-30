package reg

import (
	"fmt"
	"reflect"

	"github.com/mp3cko/registry/access"
)

// WithRegistry implementation
func withRegistryOption(useRegistry *registry) *option {
	f := func(original *registry) error {
		if !original.config.init.complete {
			return fmt.Errorf("WithRegistry used inside NewRegistry: %w", ErrNotSupported)
		}

		original.callOptions.withRegistry = useRegistry

		return nil
	}

	return newOptionWithPriority(f, priorityHighest)
}

// WithUniqueType implementation
func withUniqueTypeOption() *option {
	f := func(r *registry) error {
		if !r.config.init.complete {
			if r.config.init.uniqueTypesSet {
				return fmt.Errorf("multiple WithUniqueType calls: %w", ErrBadOption)
			}

			r.config.uniqueTypes = true
			r.config.init.uniqueTypesSet = true
			return nil
		}

		r.callOptions.uniqueType = true

		return nil
	}

	return newOption(f)
}

// WithUniqueName implementation
func withUniqueNamesOption() *option {
	f := func(r *registry) error {
		if !r.config.init.complete {
			if r.config.init.uniqueNamesSet {
				return fmt.Errorf("multiple WithUniqueName calls: %w", ErrBadOption)
			}

			r.config.uniqueNames = true
			r.config.init.uniqueNamesSet = true
			return nil
		}

		r.callOptions.uniqueName = true

		return nil
	}

	return newOption(f)
}

// WithName implementation
func withNameOption(n string) *option {
	f := func(r *registry) error {
		if !r.config.init.complete {
			if r.config.defaultName != DefaultName {
				return fmt.Errorf("WithName called multiple times: %w", ErrBadOption)
			}

			r.config.defaultName = n

			return nil
		}

		r.callOptions.name = n

		return nil
	}

	return newOption(f)
}

// WithNamedness implementation
func withNamednessOption(namedness access.Namedness) *option {
	f := func(r *registry) error {
		if !r.config.init.complete {
			if r.config.init.namednessSet {
				return fmt.Errorf("multiple WithNamedness calls: %w", ErrBadOption)
			}
			r.config.namedness = namedness
			r.config.init.namednessSet = true
		} else {
			r.callOptions.namedness = namedness
		}

		return nil
	}

	return newOptionWithPriority(f, prioritySecondHighest)
}

// WithAccessibility implementation
func withAccessibilityOption(level access.Accessibility) *option {
	f := func(r *registry) error {
		if !r.config.init.complete {
			if r.config.init.accessibilitySet {
				return fmt.Errorf("multiple WithAccessibility calls: %w", ErrBadOption)
			}
			r.config.accessibility = level
			r.config.init.accessibilitySet = true
		} else {
			r.callOptions.accessibility = level
		}

		return nil
	}

	return newOptionWithPriority(f, prioritySecondHighest)
}

// WithCloneEntries implementation
func withCloneEntriesOption(src *registry) *option {
	f := func(dest *registry) error {
		if dest.config.init.complete {
			return fmt.Errorf("WithCloneEntries used outside NewRegistry: %w", ErrNotSupported)
		}

		if err := checkBadOpt(src, dest, "WithCloneEntries"); err != nil {
			return err
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		cloneEntries(src, dest)
		// dest.config.init.clonedEntries = true

		return nil
	}

	return newOptionWithPriority(f, prioritySecondLowest)
}

// WithCloneConfig implementation
func withCloneConfigOption(src *registry) *option {
	f := func(dest *registry) error {
		if dest.config.init.complete {
			return fmt.Errorf("WithCloneConfig used outside NewRegistry: %w", ErrNotSupported)
		}

		if err := checkBadOpt(src, dest, "WithCloneConfig"); err != nil {
			return err
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		dest.config = src.config.clone()

		return nil
	}

	return newOptionWithPriority(f, priorityThirdLowest)
}

// WithCloneRegistry implementation
func withCloneRegistryOption(src *registry) *option {
	f := func(dest *registry) error {
		if dest.config.init.complete {
			return fmt.Errorf("WithCloneRegistry used outside NewRegistry: %w", ErrNotSupported)
		}

		if err := checkBadOpt(src, dest, "WithCloneRegistry"); err != nil {
			return err
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		dest.config = src.config.clone()
		cloneEntries(src, dest)
		// dest.config.init.clonedRegistry = true

		return nil
	}

	return newOptionWithPriority(f, priorityLowest)
}

func cloneEntries(src, dest *registry) {
	if dest.store == nil {
		dest.store = make(map[reflect.Type]map[string]any, len(src.store))
	}

	opts := valueOrDefault(src.callOptions, new(callOptions))

	accessibilityOption := valueOrDefault(opts.accessibility, access.AccessibilityUndefined)
	namednessOption := valueOrDefault(opts.namedness, access.NamednessUndefined)

	uniqueType := opts.uniqueType
	nameFilter := opts.name

	for rt, instances := range src.store {
		if int(namednessOption)+int(accessibilityOption) > 0 {
			rtNamedness, rtAccessibility := access.Info(rt)

			if accessibilityOption != 0 &&
				rtAccessibility != accessibilityOption ||
				namednessOption != 0 &&
					rtNamedness != namednessOption {
				continue
			}

		}
		nInstances := len(instances)
		if nInstances > 1 && uniqueType {
			continue
		}

		if nameFilter != "" {
			nInstances = 1
		}

		_, ok := dest.store[rt]
		if !ok {
			dest.store[rt] = make(map[string]any, nInstances)
		}

		for name, instance := range instances {
			if nameFilter != "" {
				if name == nameFilter {
					dest.store[rt][name] = instance
				}

				continue
			}

			if dest.config.defaultName != src.config.defaultName && name == src.config.defaultName {
				dest.store[rt][dest.config.defaultName] = instance

				continue
			}

			dest.store[rt][name] = instance
		}
	}

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

func checkBadOpt(src, dest *registry, optName string) error {
	if dest.config.init.uniqueTypesSet && src.config.init.uniqueTypesSet && dest.config.uniqueTypes != src.config.uniqueTypes {
		return fmt.Errorf("%s source WithUniqueType setting(%v) conflicts with yours(%v): %w", optName, src.config.uniqueTypes, dest.config.uniqueTypes, ErrBadOption)
	}

	if dest.config.init.uniqueNamesSet && src.config.init.uniqueNamesSet && dest.config.uniqueNames != src.config.uniqueNames {
		return fmt.Errorf("%s source WithUniqueName setting(%v) conflicts with yours(%v): %w", optName, src.config.uniqueNames, dest.config.uniqueNames, ErrBadOption)
	}

	if dest.config.accessibility != access.AccessibilityUndefined && src.config.accessibility != access.AccessibilityUndefined && dest.config.accessibility != src.config.accessibility {
		return fmt.Errorf("%s source WithAccessibility setting(%s) conflicts with yours(%s): %w", optName, src.config.accessibility, dest.config.accessibility, ErrBadOption)
	}

	if dest.config.namedness != access.NamednessUndefined && src.config.namedness != access.NamednessUndefined && dest.config.namedness != src.config.namedness {
		return fmt.Errorf("%s source WithNamedness setting(%s) conflicts with yours(%s): %w", optName, src.config.namedness, dest.config.namedness, ErrBadOption)
	}

	return nil
}
