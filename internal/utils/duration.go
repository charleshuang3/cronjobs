package utils

import (
	"fmt"
	"strings"
	"time"
)

// PrettyPrintDuration formats a time.Duration into a human-readable string, skipping zero units.
func PrettyPrintDuration(d time.Duration) string {
	d = d.Round(time.Second) // Round to the nearest second
	seconds := int(d.Seconds())
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60

	var sb strings.Builder

	if days > 0 {
		fmt.Fprintf(&sb, "%dd ", days)
	}
	if hours > 0 {
		fmt.Fprintf(&sb, "%dh ", hours)
	}
	if minutes > 0 {
		fmt.Fprintf(&sb, "%dm ", minutes)
	}
	if seconds > 0 || sb.Len() == 0 { // Always show seconds if no higher units are present
		fmt.Fprintf(&sb, "%ds", seconds)
	}

	return strings.TrimSpace(sb.String()) // Trim the space at the end
}
