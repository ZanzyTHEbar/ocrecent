package cli

import (
	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

type modelFilter struct {
	children bool
	archived bool
	dir      string
	all      bool
	extra    []string
}

func (f modelFilter) toModel() model.Filter {
	return model.Filter{
		Children: f.children,
		Archived: f.archived,
		Dir:      f.dir,
		All:      f.all,
		Extra:    f.extra,
	}
}

func palette(params *CmdParams) []*cobra.Command {
	return []*cobra.Command{
		newListCmd(params),
		newProjectsCmd(params),
		newPickCmd(params),
		newResumeCmd(params),
		newLastCmd(params),
		newPrintCmd(params),
		newNotifyCmd(params),
		newInstallCmd(params),
		newUninstallCmd(params),
		newDoctorCmd(params),
	}
}
