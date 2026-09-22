package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/format"
)

func newPickCmd(p *CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick",
		Short: "List sessions, or pick and launch with --launch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPick(p, cmd)
		},
	}
	cmd.Flags().Bool("launch", false, "launch the picked session")
	return cmd
}

func runPick(p *CmdParams, cmd *cobra.Command) error {
	if !launchFlag(cmd) {
		return runList(p, cmd)
	}
	f := filterFromFlags(cmd)
	return p.App.Pick(f.toModel(), nFromFlags(cmd), pickerFlag(cmd), projectsFlag(cmd), p.InTTY)
}

func newResumeCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "resume <id|index>",
		Short: "Resume a session by id or listing index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			return p.App.ResumeSession(args[0], f.toModel(), nFromFlags(cmd), p.InTTY)
		},
	}
}

func newLastCmd(p *CmdParams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last",
		Short: "List the most recent session, or launch it with --launch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLast(p, cmd)
		},
	}
	cmd.Flags().Bool("launch", false, "launch the most recent session")
	return cmd
}

func runLast(p *CmdParams, cmd *cobra.Command) error {
	f := filterFromFlags(cmd)
	if launchFlag(cmd) {
		return p.App.ResumeLast(f.toModel(), p.InTTY)
	}

	sessions, err := p.App.Sessions(f.toModel(), 1)
	if err != nil {
		return err
	}
	if jsonFlag(cmd) {
		out, err := format.JSONRows(sessions, nil)
		if err != nil {
			return err
		}
		fmt.Fprintln(p.Stdout, out)
		return nil
	}
	fmt.Fprintln(p.Stdout, format.SessionTable(sessions, home(), p.App.Now()))
	return nil
}

func newPrintCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "print <id|index>",
		Short: "Print the resume command for a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			s, err := p.App.Resolve(args[0], f.toModel(), nFromFlags(cmd))
			if err != nil {
				return err
			}
			fmt.Fprintln(p.Stdout, format.Print(s))
			return nil
		},
	}
}
