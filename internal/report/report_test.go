package report

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/usage"
)

func TestMarkdownSummary(t *testing.T) {
	var b bytes.Buffer
	if err := Markdown(&b, Summarize([]usage.Event{{Team: "b", CostUSD: .2}, {Team: "a", CostUSD: .1}})); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "| a | 1") || !strings.Contains(b.String(), "| b | 1") {
		t.Fatal(b.String())
	}
}
