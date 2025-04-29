package cargotoml

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// CargoToml represents the structure of Cargo.toml
type CargoToml struct {
	Package      PackageSection    `toml:"package"`
	Dependencies map[string]string `toml:"dependencies"`
}

// PackageSection represents the [package] section in Cargo.toml
type PackageSection struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Readme      string `toml:"readme"`
	Homepage    string `toml:"homepage"`
	Repository  string `toml:"repository"`
	License     string `toml:"license"`
	Edition     string `toml:"edition"`
}

func ReadBytes(content []byte) (CargoToml, error) {
	var cargo CargoToml
	if _, err := toml.Decode(string(content), &cargo); err != nil {
		return CargoToml{}, fmt.Errorf("failed to parse TOML: %w", err)
	}
	return cargo, nil
}
