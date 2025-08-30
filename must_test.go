package reg

import (
	"reflect"
	"testing"
)

func TestMustHelpers_Basic(t *testing.T) {
	r := newTestReg(t)

	MustSet(ExportedNamedTester{ID: 9}, WithRegistry(r))

	if got := MustGet[ExportedNamedTester](WithRegistry(r)); got.ID != 9 {
		t.Fatalf("MustGet got = %+v", got)
	}

	if m := MustGetAll(WithRegistry(r)); len(m) == 0 {
		t.Fatalf("MustGetAll returned empty")
	}
	// Unset must not panic on success
	MustUnset(ExportedNamedTester{}, WithRegistry(r))
}

func TestMustHelpers_PanicOnError(t *testing.T) {
	r := newTestReg(t)

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("MustGet did not panic on error")
		}
	}()

	MustGet[ExportedNamedTester](WithRegistry(r))
}

// helpers reused by multiple test files
// ExportedNamed is defined in test_helpers.go; we reuse the same type across package reg tests

func TestMustHelpers_MapIsolation(t *testing.T) {
	r := newTestReg(t)

	MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))

	m := MustGetAll(WithRegistry(r))

	rt := reflect.TypeFor[ExportedNamedTester]()

	delete(m[rt], r.config.defaultName)

	if _, err := Get[ExportedNamedTester](WithRegistry(r)); err != nil {
		t.Fatalf("Registry affected by MustGetAll map mutation: %v", err)
	}
}

func TestMustUnset_ErrorsBubbleAsPanics(t *testing.T) {
	r := newTestReg(t)

	defer func() {
		if rec := recover(); rec == nil {
			t.Fatalf("MustUnset did not panic on error")
		}
	}()

	MustUnset(ExportedNamedTester{}, WithRegistry(r))
}
