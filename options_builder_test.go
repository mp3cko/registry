package reg

import (
	"reflect"
	"testing"
)

func TestOptionsBuilder(t *testing.T) {
	testCases := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "set get equivalence",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				if err := Set(ExportedNamedTester{ID: 1}, WithRegistry(r), WithName("fn")); err != nil {
					tt.Fatalf("Set with function options error: %v", err)
				}
				if err := Set(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("builder")); err != nil {
					tt.Fatalf("Set with builder options error: %v", err)
				}
				if v, err := Get[ExportedNamedTester](WithRegistry(r), WithName("fn")); err != nil || v.ID != 1 {
					tt.Fatalf("Get with function options got=%+v err=%v", v, err)
				}
				if v, err := Get[ExportedNamedTester](WithRegistry(r).WithName("builder")); err != nil || v.ID != 2 {
					tt.Fatalf("Get with builder options got=%+v err=%v", v, err)
				}
			},
		},
		{
			name: "get all equivalence",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("builder"))
				all1 := MustGetAll(WithRegistry(r))
				all2 := MustGetAll(WithRegistry(r))
				if len(all1) == 0 || len(all2) == 0 {
					tt.Fatalf("GetAll returned empty maps")
				}
			},
		},
		{
			name: "unset equivalence",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r).WithName("fn"))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("builder"))
				if err := Unset(ExportedNamedTester{}, WithRegistry(r), WithName("fn")); err != nil {
					tt.Fatalf("Unset function options error: %v", err)
				}
				if err := Unset(ExportedNamedTester{}, WithRegistry(r).WithName("builder")); err != nil {
					tt.Fatalf("Unset builder options error: %v", err)
				}
			},
		},
		{
			name: "get all name filter equivalence",
			testFunc: func(tt *testing.T) {
				r := newTestReg(tt)
				MustSet(ExportedNamedTester{ID: 1}, WithRegistry(r))
				MustSet(ExportedNamedTester{ID: 2}, WithRegistry(r).WithName("b"))

				// function style name filter
				m1 := MustGetAll(WithRegistry(r), WithName("b"))
				// builder style name filter
				m2 := MustGetAll(WithRegistry(r).WithName("b"))
				rt := reflect.TypeFor[ExportedNamedTester]()
				if len(m1[rt]) != 1 || len(m2[rt]) != 1 {
					tt.Fatalf("GetAll name filter parity mismatch: m1=%d m2=%d", len(m1[rt]), len(m2[rt]))
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
