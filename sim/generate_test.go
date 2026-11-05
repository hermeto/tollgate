package sim

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/tokens"
)

func TestGenTextIsDeterministicAndCountsOneTokenPerPiece(t *testing.T) {
	a := genText(rand.New(rand.NewPCG(1, 2)), 200)
	b := genText(rand.New(rand.NewPCG(1, 2)), 200)
	if strings.Join(a, "") != strings.Join(b, "") {
		t.Fatal("same RNG produced different text")
	}
	if len(a) != 200 {
		t.Fatalf("pieces = %d, want 200", len(a))
	}
	if got := tokens.Count(strings.Join(a, "")); got != len(a) {
		t.Fatalf("tokens.Count(text) = %d, want %d (one token per piece)", got, len(a))
	}
	if strings.HasPrefix(a[0], " ") || a[0][:1] != strings.ToUpper(a[0][:1]) {
		t.Fatalf("first piece %q should be a capitalised word without leading space", a[0])
	}
}

func TestVocabularyWordsAreSingleTokens(t *testing.T) {
	for _, w := range vocabulary {
		if n := tokens.Count(" " + w); n != 1 {
			t.Errorf("%q counts as %d tokens", w, n)
		}
	}
}

func TestDistributionSampling(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	if v := (Distribution{Kind: "fixed", Value: 200}).sample(rng); v != 200 {
		t.Fatalf("fixed = %v", v)
	}
	lo, hi := 10.0, 20.0
	d := Distribution{Kind: "normal", Mean: 15, Stddev: 100, Min: &lo, Max: &hi}
	for range 1000 {
		if v := d.sample(rng); v < lo || v > hi {
			t.Fatalf("clamped normal produced %v", v)
		}
	}
	var sum float64
	ln := Distribution{Kind: "lognormal", Mu: 0, Sigma: 0.5}
	for range 10000 {
		v := ln.sample(rng)
		if v <= 0 {
			t.Fatalf("lognormal produced %v", v)
		}
		sum += v
	}
	// E[X] = exp(mu + sigma^2/2) ≈ 1.133 for mu=0, sigma=0.5.
	if mean := sum / 10000; mean < 1.08 || mean > 1.19 {
		t.Fatalf("lognormal mean = %.3f, want ≈ 1.133", mean)
	}
	if v := (Distribution{Kind: "normal", Mean: -50, Stddev: 0}).sample(rng); v != 0 {
		t.Fatalf("negative sample not floored at 0: %v", v)
	}
}

func TestOutputLengthRespectsMaxTokens(t *testing.T) {
	s, err := New(DefaultConfig(FlavorOpenAI), Options{})
	if err != nil {
		t.Fatal(err)
	}
	s.cfg.OutputTokens = Distribution{Kind: "fixed", Value: 50}
	rng := rand.New(rand.NewPCG(5, 6))
	if n, cut := s.outputLength(rng, 0); n != 50 || cut {
		t.Fatalf("no limit: n=%d cut=%v", n, cut)
	}
	if n, cut := s.outputLength(rng, 10); n != 10 || !cut {
		t.Fatalf("limit 10: n=%d cut=%v", n, cut)
	}
	if n, cut := s.outputLength(rng, 80); n != 50 || cut {
		t.Fatalf("limit 80: n=%d cut=%v", n, cut)
	}
}
