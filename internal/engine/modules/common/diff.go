package common

import (
	"fmt"
)

type DiffSideBySide struct {
	Path string
	Old  string
	New  string
}

func RenderSideBySide(d DiffSideBySide) string {
	return fmt.Sprintf("=== %s ===\n--- OLD ---\n%s\n--- NEW ---\n%s\n", d.Path, d.Old, d.New)
}
