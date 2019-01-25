package retry

import (
	"math"
	"time"
)

// PowBackoff uses powers of 2:  2, 4, 8, 16, 32, ...
var PowBackoff BackOffFunc = func(i int) time.Duration {
	return time.Duration(math.Pow(2, float64(i))) * time.Second
}

// TestBackoff returns a linear and super-fast backoff strategy to use in tests
var TestBackoff BackOffFunc = func(i int) time.Duration {
	return 1 * time.Microsecond
}
