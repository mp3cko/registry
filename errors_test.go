package reg

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrors_Sentinels(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"ErrNotFound", ErrNotFound},
		{"ErrNotUniqueType", ErrNotUniqueType},
		{"ErrNotUniqueName", ErrNotUniqueName},
		{"ErrNotSupported", ErrNotSupported},
		{"ErrAccessibilityTooLow", ErrAccessibilityTooLow},
		{"ErrNamednessTooLow", ErrNamednessTooLow},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			wrapped := fmt.Errorf("wrap: %w", tc.err)
			if !errors.Is(wrapped, tc.err) {
				tt.Fatalf("wrapped error not recognized by errors.Is for %v", tc.err)
			}
		})
	}
}
