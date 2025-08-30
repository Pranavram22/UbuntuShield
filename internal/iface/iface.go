package iface

import "context"

type DryRunResult struct {
	Diffs    []FileDiff
	Warnings []string
}

type FileDiff struct {
	Path string
	Old  string
	New  string
}

type Module interface {
	Name() string
	DryRun(ctx context.Context) (DryRunResult, error)
	Apply(ctx context.Context) error
	Rollback(ctx context.Context) error
}
