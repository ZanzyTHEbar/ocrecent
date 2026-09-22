package core

import (
	"strings"
	"testing"
)

func TestLauncherContentsQuoteExecutablePaths(t *testing.T) {
	service, desktop := launcherContents("/tmp/oc recent/bin/ocrecent")
	if !strings.Contains(service, `ExecStart="/tmp/oc recent/bin/ocrecent" notify`) {
		t.Fatalf("systemd launcher does not quote spaces: %q", service)
	}
	if !strings.Contains(desktop, `Exec="/tmp/oc recent/bin/ocrecent" notify`) {
		t.Fatalf("desktop launcher does not quote spaces: %q", desktop)
	}
}
