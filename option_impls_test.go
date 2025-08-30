package reg

import (
	"reflect"
	"testing"

	"github.com/mp3cko/registry/access"
)

func TestOptionImpls_NewRegistry_WithRegistry(t *testing.T) {
	_, err := NewRegistry(WithRegistry(nil))
	if err == nil {
		t.Fatalf("expected error for NewRegistry WithRegistry(), got nil")
	}

}

func TestOptionImpls_CloneEntries_NameMapping(t *testing.T) {
	src := newTestReg(t)

	MustSet(ExportedNamedTester{ID: 1}, WithRegistry(src)) // default name ""

	// dest with different default name; cloned entry under default name should remap
	dest, err := NewRegistry(WithName("foo"), WithCloneEntries(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneEntries error = %v", err)
	}

	m := MustGetAll(WithRegistry(dest))

	rt := reflect.TypeFor[ExportedNamedTester]()
	if _, ok := m[rt]["foo"]; !ok {
		t.Fatalf("cloned entry not remapped to dest default name")
	}
}

func TestOptionImpls_CloneConfig(t *testing.T) {
	src, err := NewRegistry(WithUniqueType(), WithUniqueName(), WithName("x"))
	if err != nil {
		t.Fatalf("NewRegistry src error = %v", err)
	}

	dest, err := NewRegistry(WithCloneConfig(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneConfig error = %v", err)
	}

	if !dest.config.uniqueTypes || !dest.config.uniqueNames || dest.config.defaultName != "x" {
		t.Fatalf("cloned config mismatch: %+v", dest.config)
	}
}

func TestOptionImpls_CloneConfig_OverridesEarlierName(t *testing.T) {
	src, err := NewRegistry(WithName("from-src"))
	if err != nil {
		t.Fatalf("NewRegistry src error = %v", err)
	}

	dest, err := NewRegistry(WithName("pre"), WithCloneConfig(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneConfig error = %v", err)
	}

	if dest.config.defaultName != "from-src" {
		t.Fatalf("WithCloneConfig should override prior WithName, got %q", dest.config.defaultName)
	}
}

func TestOptionImpls_CloneRegistry(t *testing.T) {
	src, err := NewRegistry(WithUniqueType(), WithName("x"))
	if err != nil {
		t.Fatalf("NewRegistry src error = %v", err)
	}

	og := ExportedNamedTester{ID: 1}
	MustSet(og, WithRegistry(src))

	dest, err := NewRegistry(WithCloneRegistry(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneRegistry error = %v", err)
	}

	if !dest.config.uniqueTypes || dest.config.defaultName != "x" {
		t.Fatalf("cloned registry config mismatch: src:%v dest:%v", src.config, dest.config)
	}

	cp := MustGet[ExportedNamedTester](WithRegistry(dest))
	if cp != og {
		t.Fatalf("cloned registry entry mismatch: got: %v, expected: %v", cp, og)
	}
}

func TestOptionImpls_CloneEntries_InvalidCombos(t *testing.T) {
	src := newTestReg(t,
		WithAccessibility(access.AccessibleEverywhere).
			WithNamedness(access.NamedType),
	)

	tcs := []struct {
		opts    []Option
		wantErr bool
		name    string
	}{
		{
			name:    "WithCloneEntries and WithUniqueType and WithAccessibility(AccessibleEverywhere)",
			opts:    []Option{WithCloneEntries(src), WithUniqueType().WithAccessibility(access.AccessibleEverywhere)},
			wantErr: false,
		},
		{
			name:    "WithCloneEntries and WithUniqueName and WithAccessibility(AccessibleEverywhere)",
			opts:    []Option{WithCloneEntries(src), WithUniqueName().WithAccessibility(access.AccessibleEverywhere)},
			wantErr: false,
		},
		{
			name:    "WithCloneEntries and WithAccessibility(AccessibleInsidePackage)",
			opts:    []Option{WithCloneEntries(src), WithAccessibility(access.AccessibleInsidePackage)},
			wantErr: true,
		},
		{
			name:    "WithCloneEntries and WithNamedness(AnonymousType)",
			opts:    []Option{WithCloneEntries(src), WithNamedness(access.AnonymousType)},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(tt *testing.T) {
			dest, err := NewRegistry(tc.opts...)
			if gotErr := err != nil; gotErr != tc.wantErr {
				if dest != nil {
					tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v dest conf: %+v", tc.wantErr, err, tc.opts, src.config, dest.config)
				}
				tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v", tc.wantErr, err, tc.opts, src.config)
			}
		})
	}

}

func TestOptionImpls_CloneConfig_InvalidCombos(t *testing.T) {
	src := newTestReg(t,
		WithAccessibility(access.AccessibleEverywhere).
			WithNamedness(access.NamedType),
	)

	tcs := []struct {
		opts    []Option
		wantErr bool
		name    string
	}{
		{
			name:    "WithCloneConfig and WithUniqueType and WithAccessibility(AccessibleEverywhere)",
			opts:    []Option{WithCloneConfig(src), WithUniqueType().WithAccessibility(access.AccessibleEverywhere)},
			wantErr: false,
		},
		{
			name:    "WithCloneConfig and WithUniqueName and WithAccessibility(AccessibleEverywhere)",
			opts:    []Option{WithCloneConfig(src), WithUniqueName().WithAccessibility(access.AccessibleEverywhere)},
			wantErr: false,
		},
		{
			name:    "WithCloneConfig and WithAccessibility(AccessibleInsidePackage)",
			opts:    []Option{WithCloneConfig(src), WithAccessibility(access.AccessibleInsidePackage)},
			wantErr: true,
		},
		{
			name:    "WithCloneConfig and WithNamedness(AnonymousType)",
			opts:    []Option{WithCloneConfig(src), WithNamedness(access.AnonymousType)},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(tt *testing.T) {
			dest, err := NewRegistry(tc.opts...)
			if gotErr := err != nil; gotErr != tc.wantErr {
				if dest != nil {
					tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v dest conf: %+v", tc.wantErr, err, tc.opts, src.config, dest.config)
				}
				tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v", tc.wantErr, err, tc.opts, src.config)
			}
		})
	}

}

func TestOptionImpls_CloneRegistry_InvalidCombos(t *testing.T) {
	src := newTestReg(t,
		WithName("TEST").
			WithAccessibility(access.NotAccessible).
			WithNamedness(access.AnonymousType).
			WithUniqueName().
			WithUniqueType(),
	)

	tcs := []struct {
		opts    []Option
		wantErr bool
		name    string
	}{
		{
			name:    "WithAccessibility(AccessibleInsidePackage) and WithCloneRegistry",
			opts:    []Option{WithAccessibility(access.AccessibleInsidePackage), WithCloneRegistry(src)},
			wantErr: true,
		},
		{
			name:    "WithAccessibility(AccessibleInsidePackage) and WithNamedness(NamednessUndefined) and WithCloneRegistry",
			opts:    []Option{WithAccessibility(access.AccessibleInsidePackage), WithNamedness(access.NamednessUndefined), WithCloneRegistry(src)},
			wantErr: true,
		},
		{
			name:    "WithCloneRegistry and WithCloneEntries and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithCloneEntries(src), WithAccessibility(access.NotAccessible)},
			wantErr: false,
		},
		{
			name:    "WithCloneRegistry and WithCloneConfig and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithCloneConfig(src), WithAccessibility(access.NotAccessible)},
			wantErr: false,
		},
		{
			name:    "WithCloneRegistry and WithRegistry and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithRegistry(src), WithAccessibility(access.NotAccessible)},
			wantErr: true,
		},
		{
			name:    "WithCloneRegistry and WithUniqueType and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithUniqueType(), WithAccessibility(access.NotAccessible)},
			wantErr: false,
		},
		{
			name:    "WithCloneRegistry and WithUniqueName and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithUniqueName(), WithAccessibility(access.NotAccessible)},
			wantErr: false,
		},
		{
			name:    "WithCloneRegistry and AccessibleEverywhere",
			opts:    []Option{WithCloneRegistry(src), WithAccessibility(access.AccessibleEverywhere)},
			wantErr: true,
		},
		{
			name:    "WithCloneRegistry and NamedType and WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithNamedness(access.NamedType), WithAccessibility(access.NotAccessible)},
			wantErr: true,
		},
		{
			name:    "WithCloneRegistry and WithName(\"x\"), WithAccessibility(NotAccessible)",
			opts:    []Option{WithCloneRegistry(src), WithName("x"), WithAccessibility(access.NotAccessible)},
			wantErr: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(tt *testing.T) {
			dest, err := NewRegistry(tc.opts...)
			if gotErr := err != nil; gotErr != tc.wantErr {
				if dest != nil {
					tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v dest conf: %+v", tc.wantErr, err, tc.opts, src.config, dest.config)
				}
				tt.Fatalf("expected error: %v, got: %v, opts: %v src conf: %+v", tc.wantErr, err, tc.opts, src.config)
			}
		})
	}

}
