package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

type ResourceUsage struct {
	CPU    string // e.g., "3.2"
	Memory string // e.g., "4500" (KB)
}

func GetResourceUsage(pid int) (ResourceUsage, error) {
	if pid <= 0 {
		return ResourceUsage{}, fmt.Errorf("invalid pid")
	}
	// ps -o %cpu,rss -p <pid>
	// Output header: %CPU RSS
	cmd := exec.Command("ps", "-o", "%cpu,rss", "-p", fmt.Sprintf("%d", pid))
	out, err := cmd.Output()
	if err != nil {
		return ResourceUsage{}, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return ResourceUsage{}, fmt.Errorf("no process found")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return ResourceUsage{}, fmt.Errorf("parse error")
	}

	return ResourceUsage{
		CPU:    fields[0] + "%",
		Memory: fields[1] + "KB", // ps RSS is usually in KB
	}, nil
}
