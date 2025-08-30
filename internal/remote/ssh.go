package remote

import (
	"context"
	"os/exec"
)

// Run over SSH using system ssh client.
func Run(ctx context.Context, userHost string, cmd []string) (string, error) {
	args := append([]string{userHost, "--"}, cmd...)
	out, err := exec.CommandContext(ctx, "ssh", args...).CombinedOutput()
	return string(out), err
}
