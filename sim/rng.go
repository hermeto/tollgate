package sim

import (
	"hash/fnv"
	"math/rand/v2"
	"sync"
)

// occurrenceLimit bounds the per-body occurrence table. When full it is cleared; sequential replays
// clear it at the same request index, so determinism is preserved.
const occurrenceLimit = 1 << 16

// seeds derives deterministic random streams for each request (spec FR-006, FR-019).
type seeds struct {
	seed uint64

	mu     sync.Mutex
	counts map[uint64]uint64
}

func newSeeds(seed uint64) *seeds {
	return &seeds{seed: seed, counts: make(map[uint64]uint64)}
}

// requestHash identifies a request by path and body.
func requestHash(path string, body []byte) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(path))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write(body)
	return h.Sum64()
}

// content returns the RNG that decides generated text and output length: a pure function of the seed
// and the request, so identical requests always produce identical content.
func (s *seeds) content(h uint64) *rand.Rand {
	return rand.New(rand.NewPCG(s.seed, h)) // #nosec G404 -- deterministic simulation, not security
}

// perRequest returns the RNG for latency and fault decisions, seeded with the seed, the request hash
// and how many times this request has been seen, so overlapping distinct requests do not interfere.
func (s *seeds) perRequest(h uint64) *rand.Rand {
	s.mu.Lock()
	if len(s.counts) >= occurrenceLimit {
		clear(s.counts)
	}
	n := s.counts[h]
	s.counts[h] = n + 1
	s.mu.Unlock()
	return rand.New(rand.NewPCG(s.seed^h, n)) // #nosec G404 -- deterministic simulation, not security
}
