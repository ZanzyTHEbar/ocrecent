package notify

import (
	"os/exec"
	"strings"

	"github.com/ZanzyTHEbar/faults-go"
)

var CodeNotify = faults.Code("ocrecent.notify.unavailable")

// Notifier adapts the package to core.NotifierPort.
type Notifier struct{}

// Send posts a libnotify notification with pick/projects/last actions and
// blocks until it is dismissed or acted on, returning the action key.
func (Notifier) Send(title, body, urgency string) (string, error) {
	bin, err := exec.LookPath("notify-send")
	if err != nil {
		return "", faults.New(CodeNotify, "notify-send not found")
	}
	cmd := exec.Command(bin, "-w", "-a", "ocrecent", "-u", urgency, "-t", "20000",
		"-A", "pick=Pick project", "-A", "projects=Open projects", "-A", "last=Open last",
		title, body)
	out, err := cmd.Output()
	action := strings.TrimSpace(string(out))
	if err != nil && action == "" {
		return "", faults.Wrap(CodeNotify, "notify-send failed", err)
	}
	return action, nil
}
