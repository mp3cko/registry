package reg

import (
	"fmt"
	"slices"
	"testing"
)

func TestOptionsBuilder_and(t *testing.T) {
	builder := &optionsBuilder{}

	opt1 := &option{optionPriority: priorityLowest}
	opt2 := &option{optionPriority: priorityHighest}

	builder.and(opt1)
	if len(builder.o) != 1 {
		t.Fatalf("expected 1 option after and(opt1) on empty optionsBuilder, got %d", len(builder.o))
	}
	if builder.o[0] != opt1 {
		t.Fatalf("expected first option to be opt1")
	}

	builder.and(opt2)
	if len(builder.o) != 2 {
		t.Fatalf("expected 2 options after and(opt2), got %d", len(builder.o))
	}
	if builder.o[1] != opt2 {
		t.Fatalf("expected second option to be opt2")
	}
}

func TestOptionsBuilder_apply(t *testing.T) {
	reg := newTestReg(t)
	builder := &optionsBuilder{}

	testName := "test_apply"
	opt := &option{
		optionFunc: func(r *registry) error {
			r.callOptions.name = testName
			return nil
		},
	}
	builder.and(opt)

	err := builder.apply(reg)
	if err != nil {
		t.Fatalf("apply() error: %v", err)
	}

	if reg.callOptions.name != testName {
		t.Fatalf("expected name to be %q, got %q", testName, reg.callOptions.name)
	}

	builder.isGlobalInstance = true
	builder.and(opt)
	err = builder.apply(reg)
	if err != nil {
		t.Fatalf("apply() error on global optionsBuilder instance: %v", err)
	}
	if len(builder.o) != 0 {
		t.Fatalf("expected global optionsBuilder instance options to be reset, got %d options", len(builder.o))
	}
}

func TestApplyOptions(t *testing.T) {
	reg := newTestReg(t)

	var applied []string
	opt1 := &option{
		optionFunc: func(r *registry) error {
			applied = append(applied, "low")
			return nil
		},
		optionPriority: priorityLowest,
	}

	opt2 := &option{
		optionFunc: func(r *registry) error {
			applied = append(applied, "high")
			return nil
		},
		optionPriority: priorityHighest,
	}

	opt3 := &option{
		optionFunc: func(r *registry) error {
			applied = append(applied, "medium")
			return nil
		},
		optionPriority: priorityThirdHighest,
	}

	err := applyOptions(reg, opt1, opt2, opt3)
	if err != nil {
		t.Fatalf("applyOptions() error: %v", err)
	}

	expected := []string{"high", "medium", "low"}
	if len(applied) != len(expected) {
		t.Fatalf("expected %d applications, got %d", len(expected), len(applied))
	}
	for i, exp := range expected {
		if applied[i] != exp {
			t.Fatalf("expected applied[%d] to be %q, got %q", i, exp, applied[i])
		}
	}

	err = applyOptions(reg)
	if err != nil {
		t.Fatalf("applyOptions() with no options error: %v", err)
	}

	errOpt := &option{
		optionFunc: func(r *registry) error {
			return fmt.Errorf("test error")
		},
	}
	err = applyOptions(reg, errOpt)
	if err == nil {
		t.Fatalf("expected error from applyOptions with failing option")
	}
}

func TestOptionSorter(t *testing.T) {
	opt1 := &option{optionPriority: priorityLowest}
	opt2 := &option{optionPriority: priorityHighest}
	opt3 := &option{optionPriority: priorityThirdHighest}
	opt4 := &option{optionPriority: priorityHighest}

	opts := []*option{opt1, opt2, opt3, opt4}

	slices.SortStableFunc(opts, optionSorter)

	if opts[0] != opt2 {
		t.Fatalf("expected first option to have highest priority")
	}
	if opts[0].optionPriority != opts[1].optionPriority {
		t.Fatalf("expected first two options to have same highest priority")
	}
	if opts[len(opts)-1] != opt1 {
		t.Fatalf("expected last option to have lowest priority")
	}

	if optionSorter(opt1, opt2) <= 0 {
		t.Fatalf("expected optionSorter(opt1,opt2) (low<high) to return positive number")
	}

	if optionSorter(opt2, opt1) >= 0 {
		t.Fatalf("expected optionSorter(opt2,opt1) (high>low) to return negative number")
	}
	if optionSorter(opt2, opt4) != 0 {
		t.Fatalf("expected optionSorter(opt2,opt4) (high==high) to return 0")
	}
}

func TestUnwrapOptions(t *testing.T) {
	builder1 := &optionsBuilder{}
	opt1 := &option{optionPriority: priorityHighest}
	builder1.and(opt1)

	builder2 := &optionsBuilder{}
	opt2 := &option{optionPriority: priorityLowest}
	opt3 := &option{optionPriority: priorityThirdHighest}
	builder2.and(opt2, opt3)

	opts := []Option{builder1, builder2}

	unwrapped := unwrapOptions(opts)

	expectedLen := 3
	if len(unwrapped) != expectedLen {
		t.Fatalf("expected %d unwrapped options, got %d", expectedLen, len(unwrapped))
	}

	if unwrapped[0] != opt1 {
		t.Fatalf("expected first unwrapped option to be opt1")
	}
	if unwrapped[1] != opt2 {
		t.Fatalf("expected second unwrapped option to be opt2")
	}
	if unwrapped[2] != opt3 {
		t.Fatalf("expected third unwrapped option to be opt3")
	}

	empty := unwrapOptions([]Option{})
	if len(empty) != 0 {
		t.Fatalf("expected len == 0 for empty slice of Option, got %d", len(empty))
	}
}
