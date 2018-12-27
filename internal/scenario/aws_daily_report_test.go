package scenario_test

import (
	"fmt"
	"testing"

	"github.com/ymgyt/costman/internal/scenario"
)

func TestStripDot(t *testing.T) {
	tests := []struct {
		amount string
		want   string
	}{
		{"1234.123456", "1234"},
		{"1234.", "1234"},
		{"1234", "1234"},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("case %d:", i), func(t *testing.T) {
			got, want := scenario.StripDot(tc.amount), tc.want
			if got != want {
				t.Errorf("StripDot(%s) = %s, want %s", tc.unit, got, want)
			}
		})
	}
}
