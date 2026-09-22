package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/format"
)

func newPickCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "pick",
		Short: "Pick a session with the picker and resume it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPick(p, cmd)
		},
	}
}

func runPick(p *CmdParams, cmd *cobra.Command) error {
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
	return &cobra.Command{
		Use:   "last",
		Short: "Resume the most recent session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			return p.App.ResumeLast(f.toModel(), p.InTTY)
		},
	}
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
