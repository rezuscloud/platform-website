// Generated helper functions for live.templ
// NOTE: This file is NOT auto-generated. Edit manually.

package sections

import (
	"math"

	"github.com/rezuscloud/platform-website/obs"
)

// shortHost returns a shortened version of a hostname for display.
func shortHost(name string) string {
	if len(name) > 20 {
		return name[:8] + ".." + name[len(name)-8:]
	}
	return name
}

// contains checks if substr is in s.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// cpuBarPct returns the CPU usage bar width (0-100) relative to node capacity.
func cpuBarPct(host obs.Host) float64 {
	if host.CPU <= 0 {
		return 0
	}
	cap := host.CPUCores
	if cap <= 0 {
		cap = 4 // legacy fallback when capacity is unknown
	}
	return math.Min((host.CPU/cap)*100, 100)
}

// ramBarPct returns the RAM usage bar width (0-100) relative to node capacity.
func ramBarPct(host obs.Host) float64 {
	if host.RAM <= 0 {
		return 0
	}
	cap := host.RAMTotal
	if cap <= 0 {
		cap = 32768 // legacy fallback (MB)
	}
	return math.Min((host.RAM/cap)*100, 100)
}
