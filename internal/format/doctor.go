package format

import (
	"fmt"
	"strings"

	"github.com/ZanzyTHEbar/ocrecent/internal/model"
)

// Doctor renders the support surface for `ocrecent doctor`.
func Doctor(info model.DoctorInfo, configPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "version:      %s\n", info.Version)
	fmt.Fprintf(&b, "config:       %s\n", configPath)
	if info.OpenCodePath != "" {
		fmt.Fprintf(&b, "opencode:     %s\n", info.OpenCodePath)
	} else {
		b.WriteString("opencode:     not found in PATH\n")
	}
	if info.OpenCodeDBPath != "" {
		fmt.Fprintf(&b, "db path:      %s\n", info.OpenCodeDBPath)
	}
	fmt.Fprintf(&b, "picker:       %s\n", info.Picker)
	fmt.Fprintf(&b, "terminal:     %s\n", info.Terminal)
	fmt.Fprintf(&b, "sessions:     %d root, %d total\n", info.RootCount, info.AllCount)

	if len(info.DBs) == 0 {
		b.WriteString("databases:    none\n")
	}
	for _, db := range info.DBs {
		cols, ok := info.Schema[db]
		if !ok {
			fmt.Fprintf(&b, "database:     %s (unreadable)\n", db)
			continue
		}
		fmt.Fprintf(&b, "database:     %s\n", db)
		fmt.Fprintf(&b, "  columns:    session(%s)\n", strings.Join(cols, ", "))
	}
	for _, w := range info.Warnings {
		fmt.Fprintf(&b, "warning:      %s\n", w)
	}
	return strings.TrimRight(b.String(), "\n")
}
