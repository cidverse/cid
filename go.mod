module github.com/cidverse/cid

//go:platform linux/amd64
//go:platform darwin/amd64

go 1.22.0

toolchain go1.23.3

require (
	github.com/ProtonMail/gopenpgp/v3 v3.0.0
	github.com/adrg/xdg v0.5.3
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cidverse/cidverseutils/ci v0.1.0
	github.com/cidverse/cidverseutils/compress v0.1.1
	github.com/cidverse/cidverseutils/containerruntime v0.1.0
	github.com/cidverse/cidverseutils/core v0.0.0-20241124224616-d11c8afe3a26
	github.com/cidverse/cidverseutils/filesystem v0.1.1
	github.com/cidverse/cidverseutils/hash v0.1.0
	github.com/cidverse/cidverseutils/network v0.1.0
	github.com/cidverse/cidverseutils/redact v0.1.0
	github.com/cidverse/cidverseutils/version v0.1.0
	github.com/cidverse/cidverseutils/zerologconfig v0.1.1
	github.com/cidverse/go-rules v0.0.0-20231112122021-075e5e6f8abc
	github.com/cidverse/go-vcs v0.0.0-20241123123700-fb65a5c04ebd
	github.com/cidverse/normalizeci v1.1.1-0.20240908141115-5bdf33fc46f7
	github.com/cidverse/repoanalyzer v0.1.1-0.20241123141326-6ec8c45aa8a1
	github.com/go-resty/resty/v2 v2.16.2
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-version v1.7.0
	github.com/in-toto/in-toto-golang v0.9.0
	github.com/jarcoal/httpmock v1.3.1
	github.com/jinzhu/configor v1.2.2
	github.com/labstack/echo/v4 v4.12.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/oriser/regroup v0.0.0-20240925165441-f6bb0e08289e
	github.com/rs/zerolog v1.33.0
	github.com/samber/lo v1.47.0
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.10.0
	gopkg.in/yaml.v3 v3.0.1
	oras.land/oras-go/v2 v2.5.0
)

require (
	cel.dev/expr v0.18.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/BurntSushi/toml v1.4.0 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.1.3 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/charlievieth/fastwalk v1.0.9 // indirect
	github.com/cidverse/cidverseutils/exec v0.1.0 // indirect
	github.com/cloudflare/circl v1.5.0 // indirect
	github.com/cyphar/filepath-securejoin v0.3.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.7 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.0 // indirect
	github.com/go-git/go-git/v5 v5.12.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/cel-go v0.22.1 // indirect
	github.com/google/go-github/v64 v64.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gosimple/slug v1.14.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.2.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pjbgf/sha1cd v0.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.8.0 // indirect
	github.com/sergi/go-diff v1.3.2-0.20230802210424-5b0b94c5c0d3 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/skeema/knownhosts v1.3.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stoewer/go-strcase v1.3.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xanzy/go-gitlab v0.114.0 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	golang.org/x/crypto v0.29.0 // indirect
	golang.org/x/exp v0.0.0-20241108190413-2d47ceb2692f // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sync v0.9.0 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/text v0.20.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241118233622-e639e219e697 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

exclude (
	github.com/sergi/go-diff v1.2.0
	github.com/sergi/go-diff v1.3.0
	github.com/sergi/go-diff v1.3.1
)
