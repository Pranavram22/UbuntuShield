package example

import (
	"context"

	"linux-hardener/internal/distro"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/policy"
)

type Example struct{}

func (e *Example) Name() string { return "ExamplePlugin" }
func (e *Example) DryRun(ctx context.Context) (iface.DryRunResult, error) {
	return iface.DryRunResult{Warnings: []string{"example plugin: no changes"}}, nil
}
func (e *Example) Apply(ctx context.Context) error { return nil }
func (e *Example) Rollback(ctx context.Context) error { return nil }

func Factory(p policy.Policy, d distro.InfoData) iface.Module { return &Example{} }
