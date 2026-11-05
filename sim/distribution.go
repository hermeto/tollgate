package sim

import (
	"math"
	"math/rand/v2"
)

// sample draws one value from the distribution, applying its clamps. Negative values become 0.
func (d Distribution) sample(rng *rand.Rand) float64 {
	var v float64
	switch d.Kind {
	case "normal":
		v = d.Mean + d.Stddev*rng.NormFloat64()
	case "lognormal":
		v = math.Exp(d.Mu + d.Sigma*rng.NormFloat64())
	default: // fixed
		v = d.Value
	}
	if d.Min != nil && v < *d.Min {
		v = *d.Min
	}
	if d.Max != nil && v > *d.Max {
		v = *d.Max
	}
	return math.Max(v, 0)
}
