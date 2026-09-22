package main

import (
	"fmt"
	"os"

	"github.com/ZanzyTHEbar/faults-go"

	"github.com/ZanzyTHEbar/ocrecent/internal/cli"
	"github.com/ZanzyTHEbar/ocrecent/internal/config"
	"github.com/ZanzyTHEbar/ocrecent/internal/core"
	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/notify"
	"github.com/ZanzyTHEbar/ocrecent/internal/picker"
	"github.com/ZanzyTHEbar/ocrecent/internal/resume"
	"github.com/ZanzyTHEbar/ocrecent/internal/store"
	"github.com/ZanzyTHEbar/ocrecent/internal/term"
)

var version = "dev"

func main() {
	cfg := config.Load("")
	inTTY := term.IsTTY(os.Stdin) && term.IsTTY(os.Stdout)

	app := core.New(
		&store.Store{},
		&picker.Picker{},
		&notify.Notifier{},
		&resume.Resumer{Terminal: cfg.Terminal},
	)

	params := &cli.CmdParams{
		Cfg:     cfg,
		App:     app,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		InTTY:   inTTY,
		Version: version,
	}

	err := cli.NewRoot(params).Execute()
	for _, w := range app.Warnings {
		fmt.Fprintf(os.Stderr, "ocrecent: warning: %s\n", w)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCode(err))
	}
}

func exitCode(err error) int {
	switch code := faults.CodeOf(err); code {
	case model.CodeNoDB, model.CodeStoreUnread:
		return 2
	case model.CodePickerCancel:
		return 130
	default:
		return 1
	}
}
