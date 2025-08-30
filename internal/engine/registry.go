package engine

import (
	"context"

	"linux-hardener/internal/distro"
	"linux-hardener/internal/engine/modules/firewall"
	"linux-hardener/internal/engine/modules/ssh"
	"linux-hardener/internal/engine/modules/sysctl"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/plugins"
	"linux-hardener/internal/policy"
)

type Context struct {
	Policy policy.Policy
	Distro distro.InfoData
}

func BuildModules(ctx Context) []iface.Module {
	mods := []iface.Module{
		ssh.New(ctx.Policy, ctx.Distro),
		firewall.New(ctx.Policy, ctx.Distro),
		sysctl.New(ctx.Policy),
	}
	mods = append(mods, plugins.Build(ctx.Policy, ctx.Distro)...)
	return mods
}

func DryRunAll(ctx Context) (iface.DryRunResult, error) {
	var agg iface.DryRunResult
	for _, m := range BuildModules(ctx) {
		res, err := m.DryRun(context.Background())
		if err != nil { return agg, err }
		agg.Diffs = append(agg.Diffs, res.Diffs...)
		agg.Warnings = append(agg.Warnings, res.Warnings...)
	}
	return agg, nil
}
