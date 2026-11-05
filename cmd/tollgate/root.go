package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hermeto/tollgate/internal/version"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "tollgate",
		Short:         "LLM gateway with platform policies",
		Long:          "Tollgate is an LLM gateway that enforces per-team policies: rate limits, budgets, PII redaction, caching, fallback and audit.",
		Version:       version.String(),
		SilenceErrors: true,
		SilenceUsage:  true,
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return usageError{fmt.Errorf("unknown command %q", args[0])}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			cmd.SetOut(cmd.ErrOrStderr())
			_ = cmd.Help()
			return usageError{fmt.Errorf("a command is required")}
		},
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error { return usageError{err} })
	root.AddCommand(newVersionCmd())
	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.String())
			return err
		},
	}
}
