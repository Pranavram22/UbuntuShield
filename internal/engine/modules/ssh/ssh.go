package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"linux-hardener/internal/backup"
	"linux-hardener/internal/distro"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/policy"
	"linux-hardener/internal/service"
)

type Module struct {
	pol   policy.Policy
	dist  distro.InfoData
	path  string
}

func New(p policy.Policy, d distro.InfoData) *Module {
	return &Module{pol: p, dist: d, path: "/etc/ssh/sshd_config"}
}

func (m *Module) Name() string { return "SSH" }

func (m *Module) desiredConfig(old string) string {
	lines := []string{}
	if m.pol.SSH.Port > 0 { lines = append(lines, fmt.Sprintf("Port %d", m.pol.SSH.Port)) }
	if m.pol.SSH.PasswordAuth { lines = append(lines, "PasswordAuthentication yes") } else { lines = append(lines, "PasswordAuthentication no") }
	prl := strings.ToLower(m.pol.SSH.PermitRootLogin)
	if prl == "" { prl = "prohibit-password" }
	lines = append(lines, "PermitRootLogin "+prl)
	return strings.Join(lines, "\n")+"\n"
}

func (m *Module) readCurrent() string {
	b, err := os.ReadFile(m.path)
	if err != nil { return "" }
	return string(b)
}

func (m *Module) DryRun(ctx context.Context) (iface.DryRunResult, error) {
	old := m.readCurrent()
	new := m.desiredConfig(old)
	if old == new {
		return iface.DryRunResult{}, nil
	}
	return iface.DryRunResult{Diffs: []iface.FileDiff{{Path: m.path, Old: old, New: new}}}, nil
}

func (m *Module) Apply(ctx context.Context) error {
	old := m.readCurrent()
	new := m.desiredConfig(old)
	if old == new { return nil }
	if os.Geteuid() != 0 { return backup.ErrNotRoot() }
	snap, err := backup.NewSnapshotDir()
	if err != nil { return err }
	if _, err := backup.BackupFile(snap, m.path); err != nil { return err }
	tmp := filepath.Join(os.TempDir(), "sshd_config.new")
	if err := os.WriteFile(tmp, []byte(new), 0o600); err != nil { return err }
	if _, err := exec.LookPath("sshd"); err == nil {
		if out, err := exec.CommandContext(ctx, "sshd", "-t", "-f", tmp).CombinedOutput(); err != nil {
			return fmt.Errorf("sshd syntax error: %v: %s", err, string(out))
		}
	}
	if err := os.Rename(tmp, m.path); err != nil { return err }
	service.Reload(ctx, "sshd", "ssh")
	return nil
}

func (m *Module) Rollback(ctx context.Context) error {
	snaps, err := backup.SnapshotList()
	if err != nil || len(snaps) == 0 { return fmt.Errorf("no snapshots found") }
	latest := snaps[len(snaps)-1]
	bpath := backup.Join(latest, m.path)
	if _, err := os.Stat(bpath); err != nil { return err }
	return backup.RestoreFile(bpath, m.path)
}
