package access

import (
	"reflect"
	"testing"
)

// test-only helper types
type privateType struct{}
type PublicType struct{}

// Composite anonymous type referencing an unexported named type
type container struct{ f privateType }

func TestInfo_NamednessAndAccessibility(t *testing.T) {
	n, a := Info(PublicType{})
	if n != NamedType || a != AccessibleEverywhere {
		t.Fatalf("PublicType => named=%v access=%v; want named=%v access=%v", n, a, NamedType, AccessibleEverywhere)
	}

	n, a = Info(privateType{})
	if n != NamedType || a != AccessibleInsidePackage {
		t.Fatalf("privateType => named=%v access=%v; want named=%v access=%v", n, a, NamedType, AccessibleInsidePackage)
	}

	n, a = Info[interface{ M() }](nil)
	if n != AnonymousType || a != AccessibleEverywhere {
		t.Fatalf("anonymous interface => named=%v access=%v; want named=%v access=%v", n, a, AnonymousType, AccessibleEverywhere)
	}

	type anon1 = struct{ X int }
	n, a = Info(anon1{})
	if n != AnonymousType || a != AccessibleEverywhere {
		t.Fatalf("anon1 => named=%v access=%v; want named=%v access=%v", n, a, AnonymousType, AccessibleEverywhere)
	}

	type anon2 = struct{ C container }
	n, a = Info(anon2{})
	if n != AnonymousType || a != AccessibleInsidePackage {
		t.Fatalf("anon2 => named=%v access=%v; want named=%v access=%v", n, a, AnonymousType, AccessibleInsidePackage)
	}
}

func Test_isExported(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"x", false},
		{"zyx", false},
		{"X", true},
		{"Σigma", true}, // non-ASCII uppercase should count as exported
		{"Ligma", true},
	}
	for _, tc := range cases {
		if got := isExported(tc.in); got != tc.want {
			t.Fatalf("isExported(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func Test_getNamedness_PrimitivesAndPointers(t *testing.T) {
	if got := getNamedness(reflect.TypeFor[int]()); got != NamedType {
		t.Fatalf("getNamedness(int) = %v, want %v", got, NamedType)
	}

	if got := getNamedness(reflect.TypeFor[*PublicType]()); got != NamedType {
		t.Fatalf("getNamedness(*PublicType) = %v, want %v", got, NamedType)
	}

	if got := getNamedness(reflect.TypeFor[struct{ Y string }]()); got != AnonymousType {
		t.Fatalf("getNamedness(anon struct) = %v, want %v", got, AnonymousType)
	}

	if got := getNamedness(reflect.TypeFor[interface{ anon() }]()); got != AnonymousType {
		t.Fatalf("getNamedness(anon interface) = %v, want %v", got, AnonymousType)
	}
}

// --- Additional tests merged from access_more_test.go ---

func TestAccessibility_String(t *testing.T) {
	cases := []struct {
		in   Accessibility
		want string
	}{
		{AccessibilityUndefined, "accessibility undefined"},
		{NotAccessible, "not accessible"},
		{AccessibleInsidePackage, "accessible inside package"},
		{AccessibleEverywhere, "accessible everywhere"},
	}
	for _, tc := range cases {
		if got := tc.in.String(); got != tc.want {
			t.Fatalf("Accessibility.String(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNamedness_String(t *testing.T) {
	cases := []struct {
		in   Namedness
		want string
	}{
		{NamednessUndefined, "namedness undefined"},
		{AnonymousType, "anonymous type"},
		{NamedType, "named type"},
	}
	for _, tc := range cases {
		if got := tc.in.String(); got != tc.want {
			t.Fatalf("Namedness.String(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func Test_extractCallerPKG(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"a/b/c.Func", "a/b/c"},
		{"a/b/c.(*Type).Method", "a/b/c"},
		{"main.main", "main"},
		{"noDotName", "noDotName"},
		{"pkg.path.with.many.dots.Func", "pkg.path.with.many.dots"},
	}

	for _, tc := range cases {
		if got := extractCallerPKG(tc.in); got != tc.want {
			t.Fatalf("extractCallerPKG(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// Local types for additional tests
type unexportedLocal struct{}
type ExportedLocal struct{}

func Test_accessibleEverywhere_Composites(t *testing.T) {
	if !accessibleEverywhere(reflect.TypeFor[int]()) {
		t.Fatalf("int should be accessible everywhere")
	}

	if !accessibleEverywhere(reflect.TypeFor[ExportedLocal]()) {
		t.Fatalf("ExportedLocal should be accessible everywhere")
	}

	if accessibleEverywhere(reflect.TypeFor[unexportedLocal]()) {
		t.Fatalf("unexportedLocal should NOT be accessible everywhere")
	}

	if !accessibleEverywhere(reflect.TypeFor[struct{ X []ExportedLocal }]()) {
		t.Fatalf("anon struct with exported slice elem should be accessible everywhere")
	}

	if accessibleEverywhere(reflect.TypeFor[struct{ U unexportedLocal }]()) {
		t.Fatalf("anon struct with unexported field should NOT be accessible everywhere")
	}

	if accessibleEverywhere(reflect.TypeFor[[]unexportedLocal]()) {
		t.Fatalf("slice of unexported should NOT be accessible everywhere")
	}

	if accessibleEverywhere(reflect.TypeFor[map[string]unexportedLocal]()) {
		t.Fatalf("map[string]unexported should NOT be accessible everywhere")
	}

	if accessibleEverywhere(reflect.TypeFor[func(unexportedLocal) unexportedLocal]()) {
		t.Fatalf("func using unexported types should NOT be accessible everywhere")
	}

	if !accessibleEverywhere(reflect.TypeFor[func(int) string]()) {
		t.Fatalf("func using predeclared types should be accessible everywhere")
	}

	type ExportedIface interface{ M() int }
	if !accessibleEverywhere(reflect.TypeFor[ExportedIface]()) {
		t.Fatalf("exported interface with predeclared method types should be accessible everywhere")
	}

	type unexportedIface interface{ M() int }
	if accessibleEverywhere(reflect.TypeFor[unexportedIface]()) {
		t.Fatalf("unexported interface should NOT be accessible everywhere due to receiver")
	}

	if accessibleEverywhere(reflect.TypeFor[interface{ M() unexportedLocal }]()) {
		t.Fatalf("interface referencing unexported type should NOT be accessible everywhere")
	}
}

func Test_accessibleFromPackage_SameVsOther(t *testing.T) {
	thisPkg := reflect.TypeFor[unexportedLocal]().PkgPath()
	otherPkg := thisPkg + "/other"

	if !accessibleFromPackage(reflect.TypeFor[unexportedLocal](), thisPkg) {
		t.Fatalf("unexportedLocal should be accessible from its own package")
	}
	if accessibleFromPackage(reflect.TypeFor[unexportedLocal](), otherPkg) {
		t.Fatalf("unexportedLocal should NOT be accessible from another package")
	}

	anon := reflect.TypeFor[struct{ U unexportedLocal }]()
	if !accessibleFromPackage(anon, thisPkg) {
		t.Fatalf("anon with unexported should be accessible from same package")
	}
	if accessibleFromPackage(anon, otherPkg) {
		t.Fatalf("anon with unexported should NOT be accessible from another package")
	}

	if !accessibleFromPackage(reflect.TypeFor[struct{ X []ExportedLocal }](), otherPkg) {
		t.Fatalf("anon with exported contents should be accessible from anywhere")
	}
}

func Test_getNamedness_InterfacesAndChannels(t *testing.T) {
	type NamedIface interface{ M() }
	if got := getNamedness(reflect.TypeFor[NamedIface]()); got != NamedType {
		t.Fatalf("getNamedness(NamedIface) = %v, want %v", got, NamedType)
	}

	if got := getNamedness(reflect.TypeFor[chan int]()); got != AnonymousType {
		t.Fatalf("getNamedness(chan int) = %v, want %v", got, AnonymousType)
	}

	if got := getNamedness(reflect.TypeFor[[3]string]()); got != AnonymousType {
		t.Fatalf("getNamedness([3]string) = %v, want %v", got, AnonymousType)
	}
}
