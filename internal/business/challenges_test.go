package business

import (
	"testing"
	"time"
)

func TestExpectedByNow(t *testing.T) {
	d := func(y int, m time.Month, day, h int) time.Time { return time.Date(y, m, day, h, 0, 0, 0, time.UTC) }
	tests := []struct {
		name string
		goal int64
		year int64
		now  time.Time
		want int64
	}{
		{"start of year", 24, 2026, d(2026, 1, 1, 0), 0},
		{"half of year", 24, 2026, d(2026, 7, 2, 12), 12},
		{"rounds down", 24, 2026, d(2026, 7, 2, 0), 11},
		{"leap year half", 24, 2028, d(2028, 7, 2, 0), 12},
		{"last second", 24, 2026, d(2027, 1, 1, 0).Add(-time.Second), 23},
		{"year over", 24, 2026, d(2027, 1, 1, 0), 24},
		{"past year", 24, 2020, d(2026, 5, 1, 0), 24},
		{"future year", 24, 2030, d(2026, 5, 1, 0), 0},
		{"non-UTC clock", 24, 2026, time.Date(2026, 7, 2, 14, 0, 0, 0, time.FixedZone("x", 2*3600)), 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expectedByNow(tt.goal, tt.year, tt.now); got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
