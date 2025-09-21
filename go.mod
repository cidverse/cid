module github.com/cidverse/cid

//go:platform linux/amd64
//go:platform darwin/amd64

go 1.24.0

toolchain go1.25.0

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/ProtonMail/gopenpgp/v3 v3.3.0
	github.com/adrg/xdg v0.5.3
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cidverse/cid-sdk-go v0.0.0-20250828204921-aaa482eac4b5
	github.com/cidverse/cidverseutils/ci v0.1.0
	github.com/cidverse/cidverseutils/compress v0.1.2-0.20250308170839-94a75eae5842
	github.com/cidverse/cidverseutils/containerruntime v0.1.1-0.20250210224234-b2040fc3a6b4
	github.com/cidverse/cidverseutils/core v0.0.0-20250911180938-73fe1e6268f7
	github.com/cidverse/cidverseutils/filesystem v0.1.2-0.20241219211714-77ae5cef4073
	github.com/cidverse/cidverseutils/hash v0.1.0
	github.com/cidverse/cidverseutils/network v0.1.0
	github.com/cidverse/cidverseutils/redact v0.1.0
	github.com/cidverse/cidverseutils/version v0.1.1-0.20250420190557-91249a22dcfe
	github.com/cidverse/cidverseutils/zerologconfig v0.1.2-0.20250329161944-cee6e2f5f53c
	github.com/cidverse/go-ptr v0.0.0-20240331160646-489e694bebbf
	github.com/cidverse/go-rules v0.0.0-20250614224628-52704bb6b812
	github.com/cidverse/go-vcs v0.0.0-20250827180914-98ba84f67766
	github.com/cidverse/go-vcsapp v0.0.0-20250918205034-e76ede0d0651
	github.com/cidverse/normalizeci v1.1.1-0.20250501143256-e98d5431deb0
	github.com/cidverse/repoanalyzer v0.1.1-0.20250430192345-f3aaf3eec7f3
	github.com/go-playground/validator/v10 v10.27.0
	github.com/go-resty/resty/v2 v2.16.5
	github.com/google/go-github/v71 v71.0.0
	github.com/google/go-github/v75 v75.0.0
	github.com/google/uuid v1.6.0
	github.com/gosimple/slug v1.15.0
	github.com/hashicorp/go-version v1.7.0
	github.com/heimdalr/dag v1.5.0
	github.com/in-toto/in-toto-golang v0.9.0
	github.com/jarcoal/httpmock v1.4.1
	github.com/jinzhu/configor v1.2.2
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v4 v4.13.4
	github.com/minio/minio-go/v7 v7.0.95
	github.com/opencontainers/image-spec v1.1.1
	github.com/oriser/regroup v0.0.0-20240925165441-f6bb0e08289e
	github.com/otiai10/copy v1.14.1
	github.com/owenrumney/go-sarif/v3 v3.2.3
	github.com/rs/zerolog v1.34.0
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.10.1
	github.com/stretchr/testify v1.11.1
	github.com/wk8/go-ordered-map/v2 v2.1.8
	gitlab.com/gitlab-org/api/client-go v0.146.0
	golang.org/x/oauth2 v0.31.0
	gopkg.in/yaml.v3 v3.0.1
	oras.land/oras-go/v2 v2.6.0
)

require (
	cel.dev/expr v0.24.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/ProtonMail/go-crypto v1.3.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/bradleyfalzon/ghinstallation/v2 v2.16.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/charlievieth/fastwalk v1.0.14 // indirect
	github.com/cidverse/cidverseutils/exec v0.1.0 // indirect
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/cyphar/filepath-securejoin v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.10 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.6.2 // indirect
	github.com/go-git/go-git/v5 v5.16.2 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/google/go-github/v72 v72.0.0 // indirect
	github.com/google/go-github/v74 v74.0.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/kevinburke/ssh_config v1.4.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pjbgf/sha1cd v0.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/samber/slog-common v0.19.0 // indirect
	github.com/samber/slog-zerolog/v2 v2.7.3 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.9.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/skeema/knownhosts v1.3.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tinylib/msgp v1.4.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/time v0.13.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
)

exclude (
	github.com/sergi/go-diff v1.2.0
	github.com/sergi/go-diff v1.3.0
	github.com/sergi/go-diff v1.3.1
)
