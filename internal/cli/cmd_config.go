package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/config"
)

func newConfigCmd(p *CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update the config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set one config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.Set(p.Cfg.Path, args[0], args[1]); err != nil {
				return err
			}
			fmt.Fprintf(p.Stdout, "wrote %s\n", p.Cfg.Path)
			return nil
		},
	})
	return cmd
}
