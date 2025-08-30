package reg

import "testing"

func Test_valueOrDefault(t *testing.T) {
	if got := valueOrDefault("", "d"); got != "d" {
		t.Fatalf("valueOrDefault empty => %q, want 'd'", got)
	}

	if got := valueOrDefault("x", "d"); got != "x" {
		t.Fatalf("valueOrDefault non-empty => %q, want 'x'", got)
	}

	if got := valueOrDefault(0, 3); got != 3 {
		t.Fatalf("valueOrDefault zero-int => %d, want 3", got)
	}

	if got := valueOrDefault(2, 3); got != 2 {
		t.Fatalf("valueOrDefault non-zero-int => %d, want 2", got)
	}

	if got := valueOrDefault(false, true); got != true {
		t.Fatalf("valueOrDefault false => %v, want true", got)
	}

	var pZ, pD *int
	if got := valueOrDefault(pZ, pD); got != pD {
		t.Fatalf("valueOrDefault p1,p2 => %v, want %v", got, pD)
	}

	pv := new(int)
	*pv = 5
	if got := valueOrDefault(pv, pD); got != pv {
		t.Fatalf("valueOrDefault p3,p2 => %v, want %v", got, pv)
	}
}

func Test_zeroValue(t *testing.T) {
	if got := zeroValue[int](); got != 0 {
		t.Fatalf("zeroValue[int]() => %d, want 0", got)
	}

	if got := zeroValue[string](); got != "" {
		t.Fatalf("zeroValue[string]() => %q, want ''", got)
	}

	if got := zeroValue[*int](); got != nil {
		t.Fatalf("zeroValue[*int]() => %v, want nil", got)
	}
}

// ExportedNamedTester is a shared test type used across tests in package reg
type ExportedNamedTester struct{ ID int }

// newTestReg creates a new registry for tests and fails the test on error
func newTestReg(t *testing.T, opts ...Option) *registry {
	t.Helper()

	r, err := NewRegistry(opts...)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	return r
}
