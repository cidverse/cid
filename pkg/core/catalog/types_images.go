package catalog

type ContainerImage struct {
	Repository string           `yaml:"repository,omitempty"`
	Image      string           `yaml:"image"`
	Provides   []ProvidedBinary `yaml:"provides"`
	Cache      []ImageCache     `yaml:"cache,omitempty"`
	Security   Security         `yaml:"security,omitempty"`
	User       string           `yaml:"user,omitempty"`
	Entrypoint *string          `yaml:"entrypoint,omitempty"`
	Certs      []ImageCerts     `yaml:"certs,omitempty"`

	Mounts []ContainerMount `yaml:"mounts,omitempty"` // Mounts
	Source ImageSource      `yaml:"source,omitempty"` // Source
}

type ProvidedBinary struct {
	Binary  string   `yaml:"binary"`
	Version string   `yaml:"version"`
	Alias   []string `yaml:"alias,omitempty"`
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

type ImageCerts struct {
	Type          string `yaml:"type"`
	ContainerPath string `yaml:"dir"`
}

type ImageSource struct {
	RegistryURL string `yaml:"registry_url"`
}
