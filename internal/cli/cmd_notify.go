package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ZanzyTHEbar/ocrecent/internal/format"
)

func newNotifyCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "notify",
		Short: "Show a notification with resume actions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			action, err := p.App.Notify(f.toModel(), nFromFlags(cmd), p.Cfg.NotifyUrgency)
			if err != nil {
				return err
			}
			return p.App.DispatchNotify(action, f.toModel(), nFromFlags(cmd), pickerFlag(cmd))
		},
	}
}

func newInstallCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install user systemd units and XDG autostart",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			written, err := p.App.Install()
			if err != nil {
				return err
			}
			for _, path := range written {
				fmt.Fprintf(p.Stdout, "wrote %s\n", path)
			}
			fmt.Fprintln(p.Stdout, "To enable:")
			fmt.Fprintln(p.Stdout, "  systemctl --user daemon-reload")
			fmt.Fprintln(p.Stdout, "  systemctl --user enable --now ocrecent-notify.timer")
			fmt.Fprintln(p.Stdout, "  systemctl --user import-environment WAYLAND_DISPLAY DISPLAY")
			fmt.Fprintln(p.Stdout, "# Wayland compositors prefer autostart:")
			fmt.Fprintln(p.Stdout, "#   exec-once = ocrecent notify")
			return nil
		},
	}
}

func newUninstallCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the user units and autostart entry",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			removed, err := p.App.Uninstall()
			if err != nil {
				return err
			}
			for _, path := range removed {
				fmt.Fprintf(p.Stdout, "removed %s\n", path)
			}
			return nil
		},
	}
}

func newDoctorCmd(p *CmdParams) *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose store, picker, and environment",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filterFromFlags(cmd)
			info, err := p.App.Doctor(f.toModel(), pickerFlag(cmd), p.Version)
			if err != nil {
				return err
			}
			fmt.Fprintln(p.Stdout, format.Doctor(info, p.Cfg.Path))
			return nil
		},
	}
}
