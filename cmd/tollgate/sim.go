package main

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/hermeto/tollgate/sim"
)

func newSimCmd() *cobra.Command {
	var (
		flavor, host, configPath string
		port                     int
		seed                     uint64
		logRequests              bool
	)
	cmd := &cobra.Command{
		Use:   "sim",
		Short: "Run a deterministic OpenAI or Anthropic provider simulator",
		Long: "Run a provider simulator that speaks the OpenAI or Anthropic wire protocol with a seeded latency\n" +
			"model and live fault injection (/_sim/faults). It never contacts a real provider.",
		Example: "  tollgate sim --flavor openai --port 9001\n  tollgate sim --flavor anthropic --config sim.yaml",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			o := sim.Overrides{Flavor: sim.Flavor(flavor), Host: host, Port: port}
			if cmd.Flags().Changed("seed") {
				o.Seed = &seed
			}
			if flavor == "" && configPath == "" {
				return usageError{fmt.Errorf("--flavor or --config is required")}
			}
			cfg, err := sim.LoadConfig(configPath, o)
			if err != nil {
				return err
			}
			logger := slog.New(slog.NewJSONHandler(cmd.ErrOrStderr(), nil))
			srv, err := sim.New(cfg, sim.Options{Logger: logger, LogRequests: logRequests})
			if err != nil {
				return err
			}
			return srv.ListenAndServe(cmd.Context())
		},
	}
	f := cmd.Flags()
	f.StringVar(&flavor, "flavor", "", "wire protocol: openai or anthropic (overrides the config file)")
	f.StringVar(&host, "host", "", "listen address (default 127.0.0.1)")
	f.IntVar(&port, "port", 0, "listen port (default 9001 for openai, 9002 for anthropic)")
	f.StringVar(&configPath, "config", "", "YAML configuration file (schemas/sim.schema.json)")
	f.Uint64Var(&seed, "seed", 42, "seed for generated content, latency and faults")
	f.BoolVar(&logRequests, "log-requests", false, "log method, path, status and duration of each request (never bodies or headers)")
	return cmd
}
