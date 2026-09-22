package resume

import (
	"os"
	"os/exec"

	"github.com/ZanzyTHEbar/faults-go"
	"golang.org/x/sys/unix"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
	"github.com/ZanzyTHEbar/ocrecent/internal/term"
)

// Resumer executes opencode either in-place (TTY) or inside a terminal.
type Resumer struct {
	Terminal string
}

// Resume invokes `opencode <dir> --session <id>`. With a TTY the current
// process is replaced; otherwise the command runs inside a terminal.
func (r *Resumer) Resume(dir, id string, inTTY bool) error {
	bin, err := exec.LookPath("opencode")
	if err != nil {
		return faults.New(model.CodeNotFound, "opencode not found in PATH")
	}
	argv := []string{bin, dir, "--session", id}
	if inTTY {
		return unix.Exec(bin, argv, os.Environ())
	}
	return r.spawnTerminal(argv)
}

// Spawn runs an arbitrary argv inside a terminal.
func (r *Resumer) Spawn(argv []string) error {
	if len(argv) == 0 {
		return faults.New(model.CodeUsage, "empty command")
	}
	bin, err := exec.LookPath(argv[0])
	if err != nil {
		return faults.New(model.CodeNotFound, "command not found", "command", argv[0])
	}
	return r.spawnTerminal(append([]string{bin}, argv[1:]...))
}

func (r *Resumer) spawnTerminal(argv []string) error {
	termBin, err := term.Detect(r.Terminal)
	if err != nil {
		return err
	}
	args := term.Command(termBin, argv)
	return unix.Exec(termBin, args, os.Environ())
}
