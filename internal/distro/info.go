package distro

import (
	"bufio"
	"os"
	"strings"
)

type InfoData struct {
	ID      string
	Name    string
	Version string
	PkgMgr  string
}

func Info() InfoData {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return InfoData{ID: "unknown", Name: "Linux", Version: "", PkgMgr: "apt"}
	}
	defer f.Close()
	m := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if idx := strings.IndexByte(line, '='); idx > 0 {
			k := line[:idx]
			v := strings.Trim(line[idx+1:], "\"")
			m[k] = v
		}
	}
	id := m["ID"]
	name := m["NAME"]
	ver := m["VERSION_ID"]
	return InfoData{ID: id, Name: name, Version: ver, PkgMgr: pkgMgrFor(id)}
}

func pkgMgrFor(id string) string {
	switch id {
	case "ubuntu", "debian":
		return "apt"
	case "fedora":
		return "dnf"
	case "centos", "rhel":
		return "yum"
	case "arch":
		return "pacman"
	default:
		return "apt"
	}
}
