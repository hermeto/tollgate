package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newProbeCmd() *cobra.Command {
	return &cobra.Command{Use: "probe <url>", Short: "Check gateway health for a container probe", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		u := strings.TrimRight(args[0], "/") + "/healthz"
		ctx := cmd.Context()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		client := &http.Client{Timeout: 2 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("probe: %s returned HTTP %d", u, resp.StatusCode)
		}
		return nil
	}}
}
