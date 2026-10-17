// Package util holds small helpers shared by the other cinema packages.
package util

import "strconv"

// ClampInt limits v to the range [lo, hi].
func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ClampFloat limits v to the range [lo, hi].
func ClampFloat(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// QueryInt reads an integer query parameter. A missing or invalid value gives def
// (not clamped); a valid one is clamped to [lo, hi].
func QueryInt(raw string, def, lo, hi int) int {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return ClampInt(n, lo, hi)
}
