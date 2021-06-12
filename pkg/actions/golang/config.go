package golang

type Config struct {
	Platform []struct {
		Goos   string `required:"true"`
		Goarch string `required:"true"`
	}
}
