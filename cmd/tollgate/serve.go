package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/internal/server"
)

func newServeCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{Use: "serve", Short: "Run the Tollgate gateway", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		s, err := server.New(cfg)
		if err != nil {
			return err
		}
		if err := s.Serve(cmd.Context()); err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return nil
	}, PreRunE: func(_ *cobra.Command, _ []string) error {
		if path == "" {
			return fmt.Errorf("--config is required")
		}
		return nil
	}}
	cmd.Flags().StringVar(&path, "config", "", "gateway YAML configuration file")
	return cmd
}
