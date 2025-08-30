package firewall

import (
	"context"
	"os/exec"
	"regexp"
	"strings"

	"linux-hardener/internal/distro"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/policy"
)

type Module struct {
	pol  policy.FirewallPolicy
	dist distro.InfoData
}

func New(p policy.Policy, d distro.InfoData) *Module {
	return &Module{pol: p.Firewall, dist: d}
}

func (m *Module) Name() string { return "Firewall" }

func (m *Module) DryRun(ctx context.Context) (iface.DryRunResult, error) {
	if !m.pol.Enabled {
		return iface.DryRunResult{Warnings: []string{"Firewall disabled in policy"}}, nil
	}
	var b strings.Builder
	if m.dist.PkgMgr == "apt" {
		b.WriteString("ufw enable\n")
		for _, a := range m.pol.Allow { b.WriteString("ufw allow "+a+"\n") }
	} else {
		for _, a := range m.pol.Allow {
			cmd := firewalldAddCmd(a)
			b.WriteString("firewall-cmd --permanent "+cmd+"\n")
		}
		b.WriteString("firewall-cmd --reload\n")
	}
	return iface.DryRunResult{Diffs: []iface.FileDiff{{Path: "firewall", Old: "", New: b.String()}}}, nil
}

func (m *Module) Apply(ctx context.Context) error {
	if !m.pol.Enabled { return nil }
	if m.dist.PkgMgr == "apt" {
		_ = exec.CommandContext(ctx, "ufw", "enable").Run()
		for _, a := range m.pol.Allow { _ = exec.CommandContext(ctx, "ufw", "allow", a).Run() }
		return nil
	}
	for _, a := range m.pol.Allow {
		cmd := firewalldAddArgs(a)
		_ = exec.CommandContext(ctx, "firewall-cmd", append([]string{"--permanent"}, cmd...)...).Run()
	}
	_ = exec.CommandContext(ctx, "firewall-cmd", "--reload").Run()
	return nil
}

func (m *Module) Rollback(ctx context.Context) error {
	// Minimal rollback: try to reload runtime config (best-effort)
	if m.dist.PkgMgr == "apt" {
		// ufw has no simple rollback; leaving as no-op
		return nil
	}
	_ = exec.CommandContext(ctx, "firewall-cmd", "--reload").Run()
	return nil
}

var portRe = regexp.MustCompile(`^\d{1,5}(/tcp|/udp)?$`)

func firewalldAddCmd(token string) string {
	if portRe.MatchString(token) {
		return "--add-port=" + token
	}
	return "--add-service=" + token
}

func firewalldAddArgs(token string) []string {
	if portRe.MatchString(token) {
		return []string{"--add-port=" + token}
	}
	return []string{"--add-service=" + token}
}
