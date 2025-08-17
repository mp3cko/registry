package reg

import (
	"fmt"
	"reflect"

	"github.com/mp3cko/registry/access"
)

// WithRegistry implementation
func withRegistryOption(useRegistry *registry) *option {
	f := func(original *registry) error {
		if !original.config.initComplete {
			return fmt.Errorf("WithRegistry used inside NewRegistry: %w", ErrNotSupported)
		}

		original.callOptions.withRegistry = useRegistry

		return nil
	}

	return newOptionWithPriority(f, priorityHighest)
}

// WithUniqueType implementation
func withUniqueTypesOption() *option {
	f := func(r *registry) error {
		if !r.config.initComplete {
			r.config.uniqueTypes = true
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
		if !r.config.initComplete {
			r.config.uniqueNames = true
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
		if !r.config.initComplete {
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
		if !r.config.initComplete {
			r.config.namedness = namedness
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
		if !r.config.initComplete {
			r.config.accessibility = level
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
		if dest.config.initComplete {
			return fmt.Errorf("WithCloneEntries used outside NewRegistry: %w", ErrNotSupported)
		}

		if dest.config.clonedConfig {
			return fmt.Errorf("WithCloneEntries called after WithCloneConfig: %w: use WithCloneRegistry instead", ErrNotSupported)
		}

		if src == nil {
			return nil
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		cloneEntries(src, dest)
		dest.config.clonedEntries = true

		return nil
	}

	return newOptionWithPriority(f, prioritySecondLowest)
}

// WithCloneConfig implementation
func withCloneConfigOption(src *registry) *option {
	f := func(dest *registry) error {
		if dest.config.initComplete {
			return fmt.Errorf("WithCloneConfig used outside NewRegistry: %w", ErrNotSupported)
		}

		if dest.config.clonedEntries {
			return fmt.Errorf("WithCloneConfig called after WithCloneEntries: %w: use WithCloneRegistry instead", ErrNotSupported)
		}

		if src == nil {
			return nil
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		*dest.config = *src.config

		return nil
	}

	return newOptionWithPriority(f, priorityThirdLowest)
}

// WithCloneRegistry implementation
func withCloneRegistryOption(src *registry) *option {
	f := func(dest *registry) error {
		if dest.config.initComplete {
			return fmt.Errorf("WithCloneRegistry used outside NewRegistry: %w", ErrNotSupported)
		}

		if dest.config.clonedRegistry {
			return fmt.Errorf("multiple WithCloneRegistry calls: %w", ErrNotSupported)
		}

		if dest.config.clonedEntries || dest.config.clonedConfig {
			return fmt.Errorf("WithCloneRegistry: %w in combination WithCloneConfig or WithCloneEntries", ErrNotSupported)
		}

		if src == nil {
			return nil
		}

		src.mu.Lock()
		defer src.mu.Unlock()

		cloneEntries(src, dest)

		dest.config = src.config.clone()
		dest.config.clonedRegistry = true

		return nil
	}

	return newOptionWithPriority(f, priorityLowest)
}

func cloneEntries(from, to *registry) {
	if to.store == nil {
		to.store = make(map[reflect.Type]map[string]any, len(from.store))
	}

	accessibilityOption := valueOrDefault(from.callOptions.accessibility, access.AccessibilityUndefined)
	namednessOption := valueOrDefault(from.callOptions.namedness, access.NamednessUndefined)

	for rt, instances := range from.store {
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
		if nInstances > 1 && from.callOptions.uniqueType {
			continue
		}

		if from.callOptions.name != "" {
			nInstances = 1
		}

		_, ok := to.store[rt]
		if !ok {
			to.store[rt] = make(map[string]any, nInstances)
		}

		for name, instance := range instances {
			if from.callOptions.name != "" {
				if name == from.callOptions.name {
					to.store[rt][name] = instance
				}

				continue
			}

			if to.config.defaultName != from.config.defaultName && name == from.config.defaultName {
				to.store[rt][to.config.defaultName] = instance

				continue
			}

			to.store[rt][name] = instance
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
