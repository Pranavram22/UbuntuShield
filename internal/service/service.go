package service

import (
	"context"
	"os/exec"
)

// Reload best-effort using common service managers.
func Reload(ctx context.Context, names ...string) {
	for _, n := range names {
		_ = exec.CommandContext(ctx, "systemctl", "reload", n).Run()
		_ = exec.CommandContext(ctx, "service", n, "reload").Run()
	}
}
