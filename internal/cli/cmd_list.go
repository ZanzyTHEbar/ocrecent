package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/format"
)

func newListCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List recent sessions",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(p, cmd)
		},
	}
}

func runList(p *CmdParams, cmd *cobra.Command) error {
	f := filterFromFlags(cmd)
	n := nFromFlags(cmd)
	if projectsFlag(cmd) {
		ps, err := p.App.Projects(f.toModel(), n)
		if err != nil {
			return err
		}
		if jsonFlag(cmd) {
			out, err := format.JSONRows(nil, ps)
			if err != nil {
				return err
			}
			fmt.Fprintln(p.Stdout, out)
			return nil
		}
		fmt.Fprintln(p.Stdout, format.ProjectTable(ps, home(), p.App.Now()))
		return nil
	}
	ss, err := p.App.Sessions(f.toModel(), n)
	if err != nil {
		return err
	}
	if jsonFlag(cmd) {
		out, err := format.JSONRows(ss, nil)
		if err != nil {
			return err
		}
		fmt.Fprintln(p.Stdout, out)
		return nil
	}
	fmt.Fprintln(p.Stdout, format.SessionTable(ss, home(), p.App.Now()))
	return nil
}

func newProjectsCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "List recent projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			n := nFromFlags(cmd)
			ps, err := p.App.Projects(f.toModel(), n)
			if err != nil {
				return err
			}
			if jsonFlag(cmd) {
				out, err := format.JSONRows(nil, ps)
				if err != nil {
					return err
				}
				fmt.Fprintln(p.Stdout, out)
				return nil
			}
			fmt.Fprintln(p.Stdout, format.ProjectTable(ps, home(), p.App.Now()))
			return nil
		},
	}
}
