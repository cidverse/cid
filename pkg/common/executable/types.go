package executable

import (
	"fmt"
)

const AnyVersionConstraint = ">= 0.0.0-0"

type PreferVersion string

const (
	PreferHighest PreferVersion = "highest"
	PreferLowest  PreferVersion = "lowest"
)

type ContainerSecurity struct {
	Capabilities []string `yaml:"capabilities,omitempty"`
	Privileged   bool     `yaml:"privileged,omitempty"`
}
type ContainerMount struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type ContainerCache struct {
	ID            string `yaml:"id" json:"id"`
	ContainerPath string `yaml:"dir" json:"dir"`
	MountType     string `yaml:"type,omitempty" json:"type,omitempty"`
}

type ContainerCerts struct {
	Type          string `yaml:"type"`
	ContainerPath string `yaml:"dir"`
}

var (
	ErrCheckingForExecutable = fmt.Errorf("error checking for executable")
	ErrExecutableNotFound    = fmt.Errorf("executable not found")
)
