package sim

import "testing"

func TestContentRNGIsPureFunctionOfRequest(t *testing.T) {
	a, b := newSeeds(1), newSeeds(1)
	h := requestHash("/v1/chat/completions", []byte(`{"x":1}`))
	if a.content(h).Uint64() != b.content(h).Uint64() {
		t.Fatal("same seed and request produced different content streams")
	}
	if newSeeds(2).content(h).Uint64() == a.content(h).Uint64() {
		t.Fatal("different seeds produced the same content stream")
	}
	other := requestHash("/v1/chat/completions", []byte(`{"x":2}`))
	if a.content(other).Uint64() == a.content(h).Uint64() {
		t.Fatal("different requests produced the same content stream")
	}
}

func TestPerRequestRNGDependsOnOccurrenceNotInterleaving(t *testing.T) {
	h1 := requestHash("/p", []byte("one"))
	h2 := requestHash("/p", []byte("two"))

	a := newSeeds(9)
	a1first := a.perRequest(h1).Uint64()
	a1second := a.perRequest(h1).Uint64()
	a2first := a.perRequest(h2).Uint64()

	// Same requests, different interleaving.
	b := newSeeds(9)
	b2first := b.perRequest(h2).Uint64()
	b1first := b.perRequest(h1).Uint64()
	b1second := b.perRequest(h1).Uint64()

	if a1first != b1first || a1second != b1second || a2first != b2first {
		t.Fatal("per-request streams changed with interleaving")
	}
	if a1first == a1second {
		t.Fatal("repeated request reused the same stream")
	}
}

func TestOccurrenceTableIsBounded(t *testing.T) {
	s := newSeeds(1)
	for i := range occurrenceLimit + 10 {
		s.perRequest(uint64(i))
	}
	if len(s.counts) > occurrenceLimit {
		t.Fatalf("occurrence table grew to %d", len(s.counts))
	}
}
