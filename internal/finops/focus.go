// Package finops renders usage events in a stable FOCUS-compatible subset.
package finops

import (
	"encoding/csv"
	"io"
	"strconv"

	"github.com/hermeto/tollgate/internal/usage"
)

func WriteCSV(w io.Writer, events []usage.Event) error {
	c := csv.NewWriter(w)
	if err := c.Write([]string{"InvoiceId", "BilledCost", "BillingCurrency", "ServiceName", "SubAccountName", "ResourceId"}); err != nil {
		return err
	}
	for _, e := range events {
		if err := c.Write([]string{e.RequestID, strconv.FormatFloat(e.CostUSD, 'f', 8, 64), "USD", e.Provider, e.Team, e.RealModel}); err != nil {
			return err
		}
	}
	c.Flush()
	return c.Error()
}
