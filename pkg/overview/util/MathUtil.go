/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import "math"

// RoundToTwoDecimals rounds a float64 value to 2 decimal places (e.g., 72.054321 -> 72.05)
func RoundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}
