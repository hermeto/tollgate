// Package pii detects and redacts synthetic, high-confidence PII patterns.
package pii

import (
	"regexp"
	"strings"
)

type Match struct {
	Type  string
	Count int
}

var patterns = map[string]*regexp.Regexp{"email": regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}`), "phone": regexp.MustCompile(`\+?[0-9][0-9 ()-]{7,}[0-9]`), "card": regexp.MustCompile(`\b(?:[0-9][ -]?){13,19}\b`)}

func Redact(input string) (string, []Match) {
	out := input
	var ms []Match
	for typ, re := range patterns {
		n := len(re.FindAllString(out, -1))
		if n > 0 {
			ms = append(ms, Match{typ, n})
			out = re.ReplaceAllString(out, "[REDACTED_"+strings.ToUpper(typ)+"]")
		}
	}
	return out, ms
}
