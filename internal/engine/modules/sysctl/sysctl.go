package sysctl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"linux-hardener/internal/backup"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/policy"
)

type Module struct {
	pol  policy.SysctlPolicy
	path string
}

func New(p policy.Policy) *Module {
	return &Module{pol: p.Sysctl, path: "/etc/sysctl.d/60-linux-hardener.conf"}
}

func (m *Module) Name() string { return "Sysctl" }

func (m *Module) desired() string {
	keys := make([]string, 0, len(m.pol.Params))
	for k := range m.pol.Params { keys = append(keys, k) }
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys { fmt.Fprintf(&b, "%s = %s\n", k, m.pol.Params[k]) }
	return b.String()
}

func (m *Module) current() string {
	b, err := os.ReadFile(m.path)
	if err != nil { return "" }
	return string(b)
}

func (m *Module) DryRun(ctx context.Context) (iface.DryRunResult, error) {
	old := m.current()
	new := m.desired()
	if old == new { return iface.DryRunResult{}, nil }
	return iface.DryRunResult{Diffs: []iface.FileDiff{{Path: m.path, Old: old, New: new}}}, nil
}

func (m *Module) Apply(ctx context.Context) error {
	old := m.current()
	new := m.desired()
	if old == new { return nil }
	if os.Geteuid() != 0 { return backup.ErrNotRoot() }
	snap, err := backup.NewSnapshotDir()
	if err != nil { return err }
	if _, err := backup.BackupFile(snap, m.path); err != nil && !os.IsNotExist(err) { return err }
	tmp := filepath.Join(os.TempDir(), "sysctl.hardener")
	if err := os.WriteFile(tmp, []byte(new), 0o644); err != nil { return err }
	if err := os.Rename(tmp, m.path); err != nil { return err }
	// reload kernel params
	if _, err := exec.LookPath("sysctl"); err == nil {
		_ = exec.CommandContext(ctx, "sysctl", "--system").Run()
	}
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
