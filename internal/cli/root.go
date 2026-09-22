package cli

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/config"
	"github.com/ZanzyTHEbar/ocrecent/internal/core"
)

type CmdParams struct {
	Cfg     config.Config
	App     *core.App
	Stdout  io.Writer
	Stderr  io.Writer
	InTTY   bool
	Version string
}

func NewRoot(params *CmdParams) *cobra.Command {
	root := &cobra.Command{
		Use:           "ocrecent",
		Short:         "List recent OpenCode sessions after a reboot",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if launchFlag(cmd) {
				return runPick(params, cmd)
			}
			return runList(params, cmd)
		},
	}

	pf := root.PersistentFlags()
	pf.IntP("max", "n", params.Cfg.N, "maximum number of rows")
	pf.Bool("all", false, "disable the row limit")
	pf.Bool("children", params.Cfg.Children, "include subagent (child) sessions")
	pf.Bool("archived", params.Cfg.Archived, "include archived sessions")
	pf.Bool("json", false, "emit JSON on stdout")
	pf.Bool("projects", false, "project grain instead of sessions")
	pf.String("picker", params.Cfg.Picker, "picker binary or auto")
	pf.String("dir", "", "only sessions in this directory (or descendants)")
	pf.StringArray("db", nil, "extra database path (repeatable)")
	root.Flags().Bool("launch", false, "launch the selected session instead of listing it")

	root.AddCommand(palette(params)...)
	return root
}

func home() string { return os.Getenv("HOME") }

func filterFromFlags(cmd *cobra.Command) (f modelFilter) {
	f.children, _ = cmd.Flags().GetBool("children")
	f.archived, _ = cmd.Flags().GetBool("archived")
	f.dir, _ = cmd.Flags().GetString("dir")
	f.all, _ = cmd.Flags().GetBool("all")
	f.extra, _ = cmd.Flags().GetStringArray("db")
	return f
}

func nFromFlags(cmd *cobra.Command) int {
	n, _ := cmd.Flags().GetInt("max")
	return n
}

func jsonFlag(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("json")
	return v
}

func projectsFlag(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("projects")
	return v
}

func pickerFlag(cmd *cobra.Command) string {
	v, _ := cmd.Flags().GetString("picker")
	return v
}

func launchFlag(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("launch")
	return v
}
