package reg

import (
	"reflect"
	"sync"

	"github.com/mp3cko/registry/access"
)

// registry is a type-safe registry where instances are registered and retrieved by type
// (and optionally by name if you want to register multiple instances of the same type).
type registry struct {
	mu sync.Mutex
	// store maps a type to a map[name]instance. Default name is an empty string.
	store       map[reflect.Type]map[string]any
	config      *registryConfig
	callOptions *callOptions
}

// registryConfig holds the configuration for the registry.
type registryConfig struct {
	init          initOpts             // initialization options
	defaultName   string               // default name for new instances
	uniqueTypes   bool                 // enforce single instance per type
	uniqueNames   bool                 // enforce unique names per type
	accessibility access.Accessibility // enforce type accessibility
	namedness     access.Namedness     // enforce type namedness
}

type initOpts struct {
	complete         bool // indicates that the registry has been fully initialized
	uniqueTypesSet   bool // indicates that the registry was initialized using WithUniqueType
	uniqueNamesSet   bool // indicates that the registry was initialized using WithUniqueName
	accessibilitySet bool // indicates that the registry was initialized using WithAccessibility
	namednessSet     bool // indicates that the registry was initialized using WithNamedness
}

// callOptions holds the options for a single call to the registry.
type callOptions struct {
	name          string               // instance name parameter
	uniqueName    bool                 // unique constraint on name
	uniqueType    bool                 // unique constraint on type
	withRegistry  *registry            // use instead of default registry
	accessibility access.Accessibility // type accessibility requirement
	namedness     access.Namedness     // type namedness requirement
}

func (t *registry) cleanup() {
	t.dropCallOpts()
	t.mu.Unlock()
}

func (t *registry) ensureCallOpts() *callOptions {
	if t.callOptions == nil {
		t.callOptions = new(callOptions)
	}

	return t.callOptions
}

func (t *registry) dropCallOpts() {
	if t.callOptions == nil {
		return
	}

	t.callOptions = nil
}

func (t *registryConfig) clone() *registryConfig {
	if t == nil {
		return nil
	}

	clone := *t
	return &clone
}

func (t *callOptions) withoutRegistry() *callOptions {
	t.withRegistry = nil

	return t
}
