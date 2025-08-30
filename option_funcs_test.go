package reg

import (
	"errors"
	"testing"

	"github.com/mp3cko/registry/access"
)

func TestOptionFuncs_Priorities_Ordering(t *testing.T) {
	r1 := newTestReg(t)
	r2 := newTestReg(t, WithAccessibility(access.AccessibleEverywhere))

	// Set into r2 default via builder; priority should apply WithRegistry first (so name applies to r2)
	if err := Set(ExportedNamedTester{ID: 10}, WithRegistry(r2).WithName("x")); err != nil {
		t.Fatalf("Set WithRegistry priority error: %v", err)
	}

	if _, err := Get[ExportedNamedTester](WithRegistry(r1).WithName("x")); err == nil {
		t.Fatalf("entry should not exist in r1")
	}

	if v, err := Get[ExportedNamedTester](WithRegistry(r2).WithName("x")); err != nil || v.ID != 10 {
		t.Fatalf("expected entry in r2: v=%+v err=%v", v, err)
	}
}

func TestOptionFuncs_UniqueType(t *testing.T) {
	r1 := newTestReg(t, WithUniqueType())
	if err := Set(ExportedNamedTester{ID: 1}, WithRegistry(r1)); err != nil {
		t.Fatalf("Set unique type err: %v", err)
	}

	if err := Set(ExportedNamedTester{ID: 2}, WithRegistry(r1)); !errors.Is(err, ErrNotUniqueType) {
		t.Fatalf("expected ErrNotUniqueType, got %v", err)
	}
}

func TestOptionFuncs_UniqueName(t *testing.T) {
	r2 := newTestReg(t, WithUniqueName())
	if err := Set(ExportedNamedTester{ID: 1}, WithRegistry(r2)); err != nil {
		t.Fatalf("Set unique name err: %v", err)
	}

	if err := Set(ExportedNamedTester{ID: 2}, WithRegistry(r2)); !errors.Is(err, ErrNotUniqueName) {
		t.Fatalf("expected ErrNotUniqueName, got %v", err)
	}
}

func TestOptionFuncs_CloneOrderingAndEntries(t *testing.T) {
	src := newTestReg(t)
	MustSet(ExportedNamedTester{ID: 5}, WithRegistry(src))

	_, err := NewRegistry(WithName("pre"), WithCloneConfig(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneConfig error: %v", err)
	}

	dest2, err := NewRegistry(WithCloneEntries(src))
	if err != nil {
		t.Fatalf("NewRegistry WithCloneEntries error: %v", err)
	}

	v, err := Get[ExportedNamedTester](WithRegistry(dest2))
	if err != nil {
		t.Fatalf("Get from dest2 after WithCloneEntries error: %v", err)
	}
	if v.ID != 5 {
		t.Fatalf("expected Get from dest2 to return ID=5, got %d", v.ID)
	}
}
