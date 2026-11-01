// Generated helper functions for live.templ
// NOTE: This file is NOT auto-generated. Edit manually.

package sections

import (
	"strings"

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
	return strings.Contains(s, substr)
}

// cpuBarPct returns the CPU bar width percent, capacity-aware.
func cpuBarPct(host obs.Host) float64 {
	if host.CPUCores <= 0 {
		return 0
	}
	p := (host.CPU / host.CPUCores) * 100
	if p > 100 {
		p = 100
	}
	return p
}

// ramBarPct returns the RAM bar width percent, capacity-aware.
func ramBarPct(host obs.Host) float64 {
	if host.RAMTotal <= 0 {
		return 0
	}
	p := (host.RAM / host.RAMTotal) * 100
	if p > 100 {
		p = 100
	}
	return p
}

// instanceCPUPct returns the per-instance CPU bar width: cores relative to a
// 1-core visual cap, clamped to 100%.
func instanceCPUPct(svc obs.Service) float64 {
	if svc.CPU <= 0 {
		return 0
	}
	if svc.CPU > 1 {
		return 100
	}
	return svc.CPU * 100
}
