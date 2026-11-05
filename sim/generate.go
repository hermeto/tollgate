package sim

import (
	"math"
	"math/rand/v2"
	"strings"
)

// vocabulary holds short common words. Every generated piece (" word", ".", ",") is exactly one
// token under internal/tokens, so the reported completion usage equals the token count of the text.
var vocabulary = [...]string{
	"the", "gateway", "team", "budget", "policy", "request", "model", "token", "cost", "limit",
	"cache", "trace", "route", "audit", "data", "user", "system", "answer", "result", "value",
	"service", "provider", "latency", "window", "stream", "event", "metric", "signal", "report", "review",
	"platform", "engineer", "product", "project", "release", "change", "test", "check", "rule", "flow",
	"is", "are", "was", "will", "can", "should", "must", "may", "has", "have",
	"and", "or", "but", "so", "then", "when", "while", "because", "if", "with",
	"for", "from", "into", "over", "under", "after", "before", "between", "within", "across",
	"fast", "slow", "small", "large", "new", "old", "safe", "clear", "simple", "stable",
	"every", "each", "some", "many", "few", "most", "all", "no", "one", "two",
	"keeps", "sends", "reads", "writes", "counts", "blocks", "allows", "records", "measures", "returns",
	"today", "again", "always", "never", "often", "only", "also", "still", "here", "there",
	"time", "day", "month", "price", "share", "level", "group", "owner", "key", "note",
}

// genText returns n pieces of deterministic filler text: words separated by spaces, with commas and
// sentence-ending periods, the first word of each sentence capitalised.
func genText(rng *rand.Rand, n int) []string {
	pieces := make([]string, 0, n)
	sentence := 0
	afterPunct := false // a word always follows punctuation, so symbols never merge into one token
	for len(pieces) < n {
		remaining := n - len(pieces)
		switch {
		case afterPunct:
		case sentence >= 6 && remaining > 1 && rng.IntN(5) == 0:
			pieces = append(pieces, ".")
			sentence, afterPunct = 0, true
			continue
		case sentence >= 3 && remaining > 2 && rng.IntN(9) == 0:
			pieces = append(pieces, ",")
			afterPunct = true
			continue
		}
		afterPunct = false
		w := vocabulary[rng.IntN(len(vocabulary))]
		if sentence == 0 {
			w = strings.ToUpper(w[:1]) + w[1:]
		}
		if len(pieces) > 0 {
			w = " " + w
		}
		pieces = append(pieces, w)
		sentence++
	}
	return pieces
}

// outputLength samples the number of completion tokens and applies the request's max-token limit.
// It reports whether the output was cut by that limit.
func (s *Server) outputLength(rng *rand.Rand, maxTokens int) (n int, truncated bool) {
	n = int(math.Round(s.cfg.OutputTokens.sample(rng)))
	if n < 1 {
		n = 1
	}
	if maxTokens > 0 && n > maxTokens {
		return maxTokens, true
	}
	return n, false
}
