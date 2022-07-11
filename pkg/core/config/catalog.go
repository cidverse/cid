package config

type Catalog struct {
	Actions map[string]Action `json:"actions"`
}

type Action struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Scope string `json:"scope"`
}
