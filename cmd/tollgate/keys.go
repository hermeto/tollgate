package main

import (
	"encoding/json"
	"time"

	"github.com/spf13/cobra"

	"github.com/hermeto/tollgate/internal/identity"
)

func newKeysCmd() *cobra.Command {
	var storePath string
	root := &cobra.Command{Use: "keys", Short: "Manage Tollgate virtual keys"}
	open := func() (*identity.Store, error) { return identity.Open(storePath) }
	var team, project, expires string
	create := &cobra.Command{Use: "create", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := open()
		if err != nil {
			return err
		}
		var at *time.Time
		if expires != "" {
			d, err := time.ParseDuration(expires)
			if err != nil {
				return err
			}
			v := time.Now().UTC().Add(d)
			at = &v
		}
		raw, r, err := s.Create(team, project, at)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"key": raw, "prefix": r.Prefix, "team": r.Team, "project": r.Project, "warning": "the full key is shown once"})
	}}
	create.Flags().StringVar(&team, "team", "", "team owner")
	create.Flags().StringVar(&project, "project", "", "project owner")
	create.Flags().StringVar(&expires, "expires", "", "Go duration, for example 2160h")
	list := &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		s, err := open()
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(s.List())
	}}
	revoke := &cobra.Command{Use: "revoke <prefix>", Args: cobra.ExactArgs(1), RunE: func(_ *cobra.Command, args []string) error {
		s, err := open()
		if err != nil {
			return err
		}
		return s.Revoke(args[0])
	}}
	rotate := &cobra.Command{Use: "rotate <prefix>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		s, err := open()
		if err != nil {
			return err
		}
		raw, r, err := s.Rotate(args[0], nil)
		if err != nil {
			return err
		}
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{"key": raw, "prefix": r.Prefix, "warning": "the full key is shown once"})
	}}
	root.PersistentFlags().StringVar(&storePath, "store", "./tollgate-keys.json", "key store path")
	root.AddCommand(create, list, revoke, rotate)
	return root
}
