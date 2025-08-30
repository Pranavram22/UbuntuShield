package policy

import (
	"errors"
	"fmt"
)

type Profile string

const (
	ProfileProd   Profile = "prod"
	ProfileDev    Profile = "dev"
	ProfileLaptop Profile = "laptop"
)

type SSHPolicy struct {
	PermitRootLogin string `yaml:"permit_root_login"` // yes/no/prohibit-password
	PasswordAuth    bool   `yaml:"password_auth"`
	Port            int    `yaml:"port"`
}

type SysctlPolicy struct {
	Params map[string]string `yaml:"params"`
}

type FirewallPolicy struct {
	Enabled bool     `yaml:"enabled"`
	Allow   []string `yaml:"allow"`
}

type Policy struct {
	Name     string            `yaml:"name"`
	Profile  Profile           `yaml:"profile"`
	SSH      SSHPolicy         `yaml:"ssh"`
	Sysctl   SysctlPolicy      `yaml:"sysctl"`
	Firewall FirewallPolicy    `yaml:"firewall"`
	Meta     map[string]string `yaml:"meta"`
}

func (p Policy) Validate() error {
	if p.Name == "" { return errors.New("name required") }
	switch p.Profile { case ProfileProd, ProfileDev, ProfileLaptop: default:
		return fmt.Errorf("invalid profile: %s", p.Profile)
	}
	if p.SSH.Port <= 0 || p.SSH.Port > 65535 { return errors.New("ssh.port invalid") }
	return nil
}
