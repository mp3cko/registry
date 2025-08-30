package reg

import (
	"testing"
)

func TestRegistryTypes_EnsureAndDropCallOpts(t *testing.T) {
	r := &registry{config: &registryConfig{}}
	if r.callOptions != nil {
		t.Fatalf("expected nil callOptions initially")
	}

	co := r.ensureCallOpts()
	if co == nil || r.callOptions == nil {
		t.Fatalf("ensureCallOpts did not initialize callOptions")
	}

	r.dropCallOpts()
	if r.callOptions != nil {
		t.Fatalf("dropCallOpts did not clear callOptions")
	}
}

func TestRegistryTypes_ConfigClone(t *testing.T) {
	cfg := &registryConfig{defaultName: "x", uniqueTypes: true, namedness: 1}
	clone := cfg.clone()

	if clone == cfg || clone == nil {
		t.Fatalf("clone should return a new non-nil pointer")
	}

	if clone.defaultName != "x" || !clone.uniqueTypes || clone.namedness != 1 {
		t.Fatalf("clone mismatch: %+v", clone)
	}
}

func TestRegistryTypes_CallOptionsWithoutRegistry(t *testing.T) {
	co := &callOptions{withRegistry: &registry{}}
	out := co.withoutRegistry()

	if out.withRegistry != nil {
		t.Fatalf("withoutRegistry should nil the withRegistry field")
	}
}
