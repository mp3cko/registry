package reg

import (
	"errors"
	"reflect"
	"testing"

	"github.com/mp3cko/registry/access"
)

func TestRegistry(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "set get default and named",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)

				v := ExportedNamedTester{ID: 1}
				if err := Set(v, WithRegistry(r)); err != nil {
					tt.Fatalf("Set default name error = %v", err)
				}

				got, err := Get[ExportedNamedTester](WithRegistry(r))
				if err != nil {
					tt.Fatalf("Get default name error = %v", err)
				}
				if got != v {
					tt.Fatalf("Get default = %+v, want %+v", got, v)
				}

				v2 := ExportedNamedTester{ID: 2}
				if err := Set(v2, WithRegistry(r).WithName("named")); err != nil {
					tt.Fatalf("Set named error = %v", err)
				}

				got2, err := Get[ExportedNamedTester](WithRegistry(r).WithName("named"))
				if err != nil {
					tt.Fatalf("Get named error = %v", err)
				}
				if got2 != v2 {
					tt.Fatalf("Get named = %+v, want %+v", got2, v2)
				}

				// unknown name should be not found
				_, err = Get[ExportedNamedTester](WithRegistry(r).WithName("missing"))
				if !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Get missing name err = %v, want ErrNotFound", err)
				}
			},
		},
		{
			name: "get not found type",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				_, err := Get[ExportedNamedTester](WithRegistry(r))
				if !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Get not found err = %v, want ErrNotFound", err)
				}
			},
		},
		{
			name: "unique type behavior",
			testFunc: func(tt *testing.T) {
				// unique at constructor
				r, err := NewRegistry(WithUniqueType())
				if err != nil {
					tt.Fatalf("NewRegistry WithUniqueType error = %v", err)
				}

				if err := Set(ExportedNamedTester{ID: 1}, WithRegistry(r)); err != nil {
					tt.Fatalf("Set first error = %v", err)
				}
				// setting same type again should fail
				if err := Set(ExportedNamedTester{ID: 2}, WithRegistry(r)); !errors.Is(err, ErrNotUniqueType) {
					tt.Fatalf("Set duplicate type err = %v, want ErrNotUniqueType", err)
				}

				// non-unique registry, but Get with WithUniqueType should error if multiple
				r2 := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r2))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r2).WithName("b"))

				_, err = Get[ExportedNamedTester](WithRegistry(r2).WithUniqueType())
				if !errors.Is(err, ErrNotUniqueType) {
					tt.Fatalf("Get WithUniqueType err = %v, want ErrNotUniqueType", err)
				}
			},
		},
		{
			name: "unique name behavior",
			testFunc: func(tt *testing.T) {
				r, err := NewRegistry(WithUniqueName())
				if err != nil {
					tt.Fatalf("NewRegistry WithUniqueName error = %v", err)
				}

				// first with default name
				if err := Set(ExportedNamedTester{ID: 1}, WithRegistry(r)); err != nil {
					tt.Fatalf("Set default name error = %v", err)
				}
				// second with same name should fail
				if err := Set(ExportedNamedTester{ID: 2}, WithRegistry(r)); !errors.Is(err, ErrNotUniqueName) {
					tt.Fatalf("Set duplicate name err = %v, want ErrNotUniqueName", err)
				}

				// using a different name is ok
				if err := Set(ExportedNamedTester{ID: 3}, WithRegistry(r).WithName("x")); err != nil {
					tt.Fatalf("Set different name error = %v", err)
				}

				// invalid usages should be NotSupported
				if _, err := Get[ExportedNamedTester](WithRegistry(r).WithUniqueName()); !errors.Is(err, ErrNotSupported) {
					tt.Fatalf("Get WithUniqueName err = %v, want ErrNotSupported", err)
				}
				if _, err := GetAll(WithRegistry(r).WithUniqueName()); !errors.Is(err, ErrNotSupported) {
					tt.Fatalf("GetAll WithUniqueName err = %v, want ErrNotSupported", err)
				}
				if err := Unset(ExportedNamedTester{}, WithRegistry(r).WithUniqueName()); !errors.Is(err, ErrNotSupported) {
					tt.Fatalf("Unset WithUniqueName err = %v, want ErrNotSupported", err)
				}
			},
		},
		{
			name: "unset basic and errors",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)

				// unset missing type
				if err := Unset(ExportedNamedTester{}, WithRegistry(r)); !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Unset missing type err = %v, want ErrNotFound", err)
				}

				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("b"))

				// WithUniqueType when multiple instances exist should fail
				if err := Unset(ExportedNamedTester{}, WithRegistry(r).WithUniqueType()); !errors.Is(err, ErrNotUniqueType) {
					tt.Fatalf("Unset WithUniqueType err = %v, want ErrNotUniqueType", err)
				}

				// unset named instance
				if err := Unset(ExportedNamedTester{}, WithRegistry(r).WithName("b")); err != nil {
					tt.Fatalf("Unset named error = %v", err)
				}

				// unset default instance should remove type entirely now
				if err := Unset(ExportedNamedTester{}, WithRegistry(r)); err != nil {
					tt.Fatalf("Unset default error = %v", err)
				}

				if _, err := Get[ExportedNamedTester](WithRegistry(r)); !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Get after unset err = %v, want ErrNotFound", err)
				}
			},
		},
		{
			name: "get all basic and isolation",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("b"))

				all := MustGetAll(WithRegistry(r))
				rt := reflect.TypeFor[ExportedNamedTester]()
				if got := len(all[rt]); got != 2 {
					tt.Fatalf("GetAll count = %d, want 2", got)
				}

				// Mutating the returned map must not affect the registry
				delete(all[rt], "")
				// original registry should still have default instance
				if _, err := Get[ExportedNamedTester](WithRegistry(r)); err != nil {
					tt.Fatalf("Registry affected by GetAll map mutation: %v", err)
				}
			},
		},
		{
			name: "with registry overrides",
			testFunc: func(tt *testing.T) {
				r1 := newTestReg(tt)
				r2 := newTestReg(tt)

				// Set into r2 using WithRegistry; default registry shouldn't see it
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r2).WithName("x"))

				if _, err := Get[ExportedNamedTester](WithName("x")); !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Default registry should not see r2 entry, err = %v", err)
				}

				got, err := Get[ExportedNamedTester](WithRegistry(r2).WithName("x"))
				if err != nil || got.ID != 2 {
					tt.Fatalf("Get from r2 error = %v, got = %+v", err, got)
				}

				// Get using WithRegistry from r1 should fail when empty and confirm isolation from r2
				if _, err := Get[ExportedNamedTester](WithRegistry(r1).WithName("x")); !errors.Is(err, ErrNotFound) {
					tt.Fatalf("Get from r1 err = %v, want ErrNotFound", err)
				}
			},
		},
		{
			name: "set default registry switch",
			testFunc: func(tt *testing.T) {
				// save and restore default registry at end
				orig := defReg.Load()
				r := newTestReg(tt)
				SetDefaultRegistry(r)
				defer SetDefaultRegistry(orig)

				if err := Set(ExportedNamedTester{ID: 42}); err != nil {
					tt.Fatalf("Set on new default error = %v", err)
				}
				got, err := Get[ExportedNamedTester]()
				if err != nil || got.ID != 42 {
					tt.Fatalf("Get from new default error = %v, got = %+v", err, got)
				}
			},
		},
		{
			name: "call options reset between calls",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("b"))

				// first call with name
				if _, err := Get[ExportedNamedTester](WithRegistry(r).WithName("b")); err != nil {
					tt.Fatalf("Get named error = %v", err)
				}
				// second call without name must not reuse last name
				got, err := Get[ExportedNamedTester](WithRegistry(r))
				if err != nil || got.ID != 1 {
					tt.Fatalf("Get default after named call error = %v, got = %+v", err, got)
				}
			},
		},
		{
			name: "accessibility and namedness on set",
			testFunc: func(tt *testing.T) {
				// Accessibility: require everywhere, but unexported type is only package-visible => should fail
				r, err := NewRegistry(WithAccessibility(access.AccessibleEverywhere))
				if err != nil {
					tt.Fatalf("NewRegistry WithAccessibility error = %v", err)
				}

				type unexportedType struct{ ID int }
				if err := Set(unexportedType{}, WithRegistry(r)); !errors.Is(err, ErrAccessibilityTooLow) {
					tt.Fatalf("Set unexported under AccessibleEverywhere err = %v, want ErrAccessibilityTooLow", err)
				}

				// Namedness: require named type; anonymous interface should fail
				r2, err := NewRegistry(WithNamedness(access.NamedType))
				if err != nil {
					tt.Fatalf("NewRegistry WithNamedness error = %v", err)
				}

				if err := Set[interface{ M() }](nil, WithRegistry(r2)); !errors.Is(err, ErrNamednessTooLow) {
					tt.Fatalf("Set anonymous interface err = %v, want ErrNamednessTooLow", err)
				}

				// named type should work; use exported named type to decouple from accessibility heuristics
				if err := Set(ExportedNamedTester{}, WithRegistry(r2)); err != nil {
					tt.Fatalf("Set named type error = %v", err)
				}
			},
		},
		{
			name: "default name in constructor applies to set and get",
			testFunc: func(tt *testing.T) {
				r, err := NewRegistry(WithName("def"))
				if err != nil {
					tt.Fatalf("NewRegistry WithName error = %v", err)
				}
				v := ExportedNamedTester{ID: 7}
				if err := Set(v, WithRegistry(r)); err != nil {
					tt.Fatalf("Set without name in named registry error = %v", err)
				}
				got, err := Get[ExportedNamedTester](WithRegistry(r))
				if err != nil || got != v {
					tt.Fatalf("Get without name in named registry error = %v, got = %+v", err, got)
				}
				all := MustGetAll(WithRegistry(r))
				rt := reflect.TypeFor[ExportedNamedTester]()
				if _, ok := all[rt]["def"]; !ok {
					tt.Fatalf("expected instance to be stored under default name 'def'")
				}
				if _, ok := all[rt][""]; ok {
					tt.Fatalf("did not expect empty-name entry when default name is set")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			tc.testFunc(tt)
		})
	}
}
