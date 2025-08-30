package policy

import (
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func Load(path string) (Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil { return Policy{}, err }
	var p Policy
	if err := yaml.Unmarshal(b, &p); err != nil { return Policy{}, err }
	return p, nil
}

func Save(p Policy, w io.Writer) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(p)
}
