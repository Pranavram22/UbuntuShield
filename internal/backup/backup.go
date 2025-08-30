package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func SnapshotRoot() string {
	return "/var/backups/linux-hardener"
}

func NewSnapshotDir() (string, error) {
	root := SnapshotRoot()
	if err := os.MkdirAll(root, 0o755); err != nil { return "", err }
	ts := time.Now().Format("20060102-150405")
	path := filepath.Join(root, ts)
	if err := os.MkdirAll(path, 0o755); err != nil { return "", err }
	return path, nil
}

func BackupFile(dstDir, src string) (string, error) {
	b := filepath.Join(dstDir, filepath.Base(src))
	in, err := os.ReadFile(src)
	if err != nil { return "", err }
	if err := os.WriteFile(b, in, 0o600); err != nil { return "", err }
	return b, nil
}

func RestoreFile(src, dst string) error {
	b, err := os.ReadFile(src)
	if err != nil { return err }
	return os.WriteFile(dst, b, 0o600)
}

func SnapshotList() ([]string, error) {
	root := SnapshotRoot()
	d, err := os.ReadDir(root)
	if err != nil { return nil, err }
	out := make([]string, 0, len(d))
	for _, e := range d { if e.IsDir() { out = append(out, filepath.Join(root, e.Name())) } }
	return out, nil
}

func Join(dir, file string) string { return filepath.Join(dir, filepath.Base(file)) }

func ErrNotRoot() error { return fmt.Errorf("requires root privileges") }
