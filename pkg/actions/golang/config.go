package golang

var Config = struct {
	GoLang struct {
		Platform []struct {
			Goos string `required:"true"`
			Goarch string `required:"true"`
		}
	}
}{}
