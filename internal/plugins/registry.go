package plugins

import (
	"linux-hardener/internal/distro"
	"linux-hardener/internal/iface"
	"linux-hardener/internal/policy"
)

type Factory func(p policy.Policy, d distro.InfoData) iface.Module

var factories []Factory

func Register(f Factory) { factories = append(factories, f) }

func Build(p policy.Policy, d distro.InfoData) []iface.Module {
	mods := make([]iface.Module, 0, len(factories))
	for _, f := range factories { mods = append(mods, f(p, d)) }
	return mods
}
