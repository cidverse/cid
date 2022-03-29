package golang

type Config struct {
	Debug    bool `required:"true" default:"false"`
	Platform []struct {
		Goos   string `required:"true"`
		Goarch string `required:"true"`
	}
}
