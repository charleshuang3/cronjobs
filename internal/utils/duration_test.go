package utils

import (
	"testing"
	"time"
)

func TestPrettyPrintDuration(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "0s"},
		{time.Second, "1s"},
		{2 * time.Second, "2s"},
		{time.Minute, "1m"},
		{2 * time.Minute, "2m"},
		{time.Minute + time.Second, "1m 1s"},
		{2*time.Minute + time.Second, "2m 1s"},
		{time.Hour, "1h"},
		{2 * time.Hour, "2h"},
		{time.Hour + time.Minute, "1h 1m"},
		{2*time.Hour + 3*time.Minute + 5*time.Second, "2h 3m 5s"},
		{24 * time.Hour, "1d"},
		{48 * time.Hour, "2d"},
		{10*time.Hour + 2*time.Minute + 1*time.Second, "10h 2m 1s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			output := PrettyPrintDuration(tt.input)
			if output != tt.expected {
				t.Errorf("prettyPrintDuration(%v) = %v; expected %v", tt.input, output, tt.expected)
			}
		})
	}
}
