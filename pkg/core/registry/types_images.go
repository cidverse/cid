package registry

type ContainerImage struct {
	Repository string           `yaml:"repository,omitempty"`
	Image      string           `yaml:"image"`
	Provides   []ProvidedBinary `yaml:"provides"`
	Cache      []ImageCache     `yaml:"cache,omitempty"`
	Security   Security         `yaml:"security,omitempty"`
	User       string           `yaml:"user,omitempty"`

	Mounts []ContainerMount `yaml:"mounts,omitempty"` // Mounts
	Source ImageSource      `yaml:"source,omitempty"` // Source
}

type ProvidedBinary struct {
	Binary  string `yaml:"binary"`
	Version string `yaml:"version"`
}

type Security struct {
	Capabilities []string `yaml:"capabilities,omitempty"`
	Privileged   bool     `yaml:"privileged,omitempty"`
}
type ContainerMount struct {
	Src  string `yaml:"src"`
	Dest string `yaml:"dest"`
}

type ImageCache struct {
	ID            string `yaml:"id"`
	ContainerPath string `yaml:"dir"`
	MountType     string `yaml:"type,omitempty"`
}

type ImageSource struct {
	RegistryURL string `yaml:"registry_url"`
}
