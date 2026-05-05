package obs

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSparkline(t *testing.T) {
	t.Run("empty points returns empty strings", func(t *testing.T) {
		line, area := Sparkline([]float64{}, 100, 24)
		assert.Equal(t, "", line)
		assert.Equal(t, "", area)
	})

	t.Run("single point returns flat line", func(t *testing.T) {
		line, area := Sparkline([]float64{50}, 100, 24)
		assert.Contains(t, line, ",")
		assert.Contains(t, area, ",")
	})

	t.Run("line has correct number of points", func(t *testing.T) {
		points := []float64{10, 20, 30, 40, 50}
		line, _ := Sparkline(points, 100, 24)
		// 5 points means 5 coordinate pairs separated by spaces
		pairs := splitPoints(line)
		assert.Equal(t, 5, len(pairs))
	})

	t.Run("values are scaled to height", func(t *testing.T) {
		points := []float64{0, 100}
		line, _ := Sparkline(points, 100, 24)
		// First point at bottom (y=24), second at top (y=0)
		pairs := splitPoints(line)
		assert.Equal(t, "0.0,24.0", pairs[0])
		assert.Equal(t, "100.0,0.0", pairs[1])
	})

	t.Run("area polygon closes to baseline", func(t *testing.T) {
		points := []float64{10, 50, 90}
		_, area := Sparkline(points, 100, 24)
		// Area should start at bottom-left, go through points, end at bottom-right
		assert.Contains(t, area, "0,24")   // starts at baseline
		assert.Contains(t, area, "100,24") // ends at baseline
	})

	t.Run("constant values produce flat line", func(t *testing.T) {
		points := []float64{50, 50, 50, 50}
		line, _ := Sparkline(points, 100, 24)
		pairs := splitPoints(line)
		// All Y values should be the same
		for _, p := range pairs {
			assert.Contains(t, p, ",24.0") // flat at baseline for constant values
		}
	})
}

func TestSparklineEdgeCases(t *testing.T) {
	t.Run("all zeros produces flat baseline", func(t *testing.T) {
		points := []float64{0, 0, 0, 0}
		line, _ := Sparkline(points, 100, 24)
		pairs := splitPoints(line)
		for _, p := range pairs {
			assert.Contains(t, p, ",24")
		}
	})

	t.Run("negative values are clamped to zero", func(t *testing.T) {
		points := []float64{-10, 0, 10}
		line, _ := Sparkline(points, 100, 24)
		pairs := splitPoints(line)
		assert.Equal(t, 3, len(pairs))
	})

	t.Run("NaN values are treated as zero", func(t *testing.T) {
		points := []float64{10, math.NaN(), 30}
		line, _ := Sparkline(points, 100, 24)
		assert.NotEmpty(t, line)
	})
}

// splitPoints splits "x1,y1 x2,y2 ..." into ["x1,y1", "x2,y2", ...]
func splitPoints(s string) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
