package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hermeto/tollgate/internal/policy"
)

func newPolicyCmd() *cobra.Command {
	var dir, team, project string
	root := &cobra.Command{Use: "policy", Short: "Validate and explain declarative policies"}
	validate := &cobra.Command{Use: "validate", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		if _, err := policy.Load(dir); err != nil {
			return err
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), "policy valid")
		return err
	}}
	explain := &cobra.Command{Use: "explain", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		snap, err := policy.Load(dir)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(snap.Effective(team, project))
	}}
	root.PersistentFlags().StringVar(&dir, "dir", "policies/examples", "policy directory")
	explain.Flags().StringVar(&team, "team", "", "team name")
	explain.Flags().StringVar(&project, "project", "", "project name")
	root.AddCommand(validate, explain)
	return root
}
