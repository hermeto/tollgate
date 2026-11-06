// Package report renders reproducible summaries from usage events.
package report

import (
	"fmt"
	"io"
	"sort"

	"github.com/hermeto/tollgate/internal/usage"
)

type Summary struct {
	Team     string
	Requests int
	CostUSD  float64
}

func Summarize(events []usage.Event) []Summary {
	m := map[string]*Summary{}
	for _, e := range events {
		s := m[e.Team]
		if s == nil {
			s = &Summary{Team: e.Team}
			m[e.Team] = s
		}
		s.Requests++
		s.CostUSD += e.CostUSD
	}
	out := make([]Summary, 0, len(m))
	for _, s := range m {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Team < out[j].Team })
	return out
}
func Markdown(w io.Writer, summaries []Summary) error {
	if _, err := fmt.Fprintln(w, "| Team | Requests | Cost USD |\n|---|---:|---:|"); err != nil {
		return err
	}
	for _, s := range summaries {
		if _, err := fmt.Fprintf(w, "| %s | %d | %.8f |\n", s.Team, s.Requests, s.CostUSD); err != nil {
			return err
		}
	}
	return nil
}
