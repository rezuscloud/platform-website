package obs

import (
	"fmt"
	"math"
)

// Sparkline generates SVG polyline points and area polygon points from a slice of float64 values.
// w is the SVG width, h is the SVG height. Returns (linePoints, areaPoints).
func Sparkline(values []float64, w, h int) (string, string) {
	if len(values) == 0 {
		return "", ""
	}

	// Find min/max for scaling
	min, max := values[0], values[0]
	for _, v := range values {
		if math.IsNaN(v) {
			continue
		}
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Clamp negative min to 0
	if min < 0 {
		min = 0
	}

	// Avoid division by zero
	range_ := max - min
	if range_ == 0 {
		range_ = 1
	}

	n := len(values)
	xStep := float64(w) / float64(n-1)
	if n == 1 {
		xStep = 0
	}

	// Build polyline points
	linePoints := ""
	areaPoints := fmt.Sprintf("0,%d ", h) // start at bottom-left
	for i, v := range values {
		if math.IsNaN(v) {
			v = min
		}
		if v < 0 {
			v = 0
		}
		x := float64(i) * xStep
		y := float64(h) - ((v - min) / range_ * float64(h))
		// Clamp y to SVG bounds
		if y < 0 {
			y = 0
		}
		if y > float64(h) {
			y = float64(h)
		}

		pt := fmt.Sprintf("%.1f,%.1f", x, y)
		if linePoints != "" {
			linePoints += " "
		}
		linePoints += pt

		if areaPoints != "" && i > 0 {
			areaPoints += " "
		}
		areaPoints += pt
	}
	areaPoints += fmt.Sprintf(" %d,%d", w, h) // close to bottom-right

	return linePoints, areaPoints
}
