package finops

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/usage"
)

func TestWriteCSV(t *testing.T) {
	var b bytes.Buffer
	if err := WriteCSV(&b, []usage.Event{{RequestID: "r", CostUSD: .1, Provider: "openai", Team: "payments", RealModel: "m"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "BilledCost") || !strings.Contains(b.String(), "0.10000000") {
		t.Fatal(b.String())
	}
}
