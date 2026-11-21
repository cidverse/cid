module github.com/cidverse/cid

//go:platform linux/amd64
//go:platform darwin/amd64

go 1.27.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/ProtonMail/gopenpgp/v3 v3.5.2
	github.com/adrg/xdg v0.5.3
	github.com/bwmarrin/snowflake v0.3.0
	github.com/cidverse/cidverseutils/ci v0.1.0
	github.com/cidverse/cidverseutils/compress v0.1.2-0.20250308170839-94a75eae5842
	github.com/cidverse/cidverseutils/containerruntime v0.1.1-0.20250210224234-b2040fc3a6b4
	github.com/cidverse/cidverseutils/core v0.0.0-20261008211605-afb6aea7d334
	github.com/cidverse/cidverseutils/filesystem v0.1.2-0.20260120202941-ba7eef6381fa
	github.com/cidverse/cidverseutils/hash v0.1.0
	github.com/cidverse/cidverseutils/network v0.1.0
	github.com/cidverse/cidverseutils/redact v0.1.0
	github.com/cidverse/cidverseutils/version v0.1.1-0.20250420190557-91249a22dcfe
	github.com/cidverse/cidverseutils/zerologconfig v0.1.2-0.20250329161944-cee6e2f5f53c
	github.com/cidverse/go-ptr v0.0.0-20240331160646-489e694bebbf
	github.com/cidverse/go-rules v0.0.0-20250614224628-52704bb6b812
	github.com/cidverse/go-vcs v0.0.0-20261101171423-912056c1c70a
	github.com/cidverse/go-vcsapp v0.0.0-20261115140459-ce2ad580b686
	github.com/cidverse/normalizeci v1.1.1-0.20261021160637-7dfca485a24e
	github.com/cidverse/repoanalyzer v0.1.1-0.20261101132110-946252f4aa3b
	github.com/go-playground/validator/v10 v10.30.6
	github.com/go-resty/resty/v2 v2.17.2
	github.com/google/go-github/v93 v93.0.0
	github.com/google/uuid v1.6.0
	github.com/gosimple/slug v1.15.0
	github.com/hashicorp/go-version v1.9.0
	github.com/heimdalr/dag v1.5.1
	github.com/in-toto/attestation v1.2.0
	github.com/jarcoal/httpmock v1.4.2
	github.com/jinzhu/configor v1.2.2
	github.com/joho/godotenv v1.5.1
	github.com/labstack/echo/v5 v5.4.7
	github.com/minio/minio-go/v7 v7.4.1
	github.com/opencontainers/image-spec v1.1.1
	github.com/oriser/regroup v0.0.0-20240925165441-f6bb0e08289e
	github.com/otiai10/copy v1.14.1
	github.com/owenrumney/go-sarif/v3 v3.3.2
	github.com/rs/zerolog v1.35.1
	github.com/sigstore/sigstore-go v1.4.0
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.12.1
	github.com/wk8/go-ordered-map/v2 v2.1.8
	gitlab.com/gitlab-org/api/client-go/v3 v3.54.1
	go.yaml.in/yaml/v4 v4.0.0
	golang.org/x/net v0.63.0
	golang.org/x/oauth2 v0.38.0
	golang.org/x/sync v0.25.0
	golang.org/x/sys v0.50.0
	google.golang.org/protobuf v1.36.12
	oras.land/oras-go/v2 v2.7.0
)

require (
	cel.dev/expr v0.26.0 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Microsoft/go-winio v0.6.4 // indirect
	github.com/ProtonMail/go-crypto v1.5.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/bradleyfalzon/ghinstallation/v2 v2.19.0 // indirect
	github.com/buger/jsonparser v1.6.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charlievieth/fastwalk v1.0.14 // indirect
	github.com/cidverse/cidverseutils/exec v0.1.0 // indirect
	github.com/cloudflare/circl v1.6.5 // indirect
	github.com/coreos/go-oidc/v3 v3.21.0 // indirect
	github.com/cyberphone/json-canonicalization v0.0.0-20241213102144-19d51d7fe467 // indirect
	github.com/cyphar/filepath-securejoin v0.7.0 // indirect
	github.com/digitorus/pkcs7 v0.0.0-20260821105541-05f79448fa77 // indirect
	github.com/digitorus/timestamp v0.0.0-20231217203849-220c5c2851b7 // indirect
	github.com/dustin/go-humanize v1.1.0 // indirect
	github.com/emirpasic/gods v1.18.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-git/gcfg v1.5.1-0.20230307220236-3a3c6141e376 // indirect
	github.com/go-git/go-billy/v5 v5.9.2 // indirect
	github.com/go-git/go-git/v5 v5.19.3 // indirect
	github.com/go-jose/go-jose/v4 v4.1.6 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/analysis v1.0.0 // indirect
	github.com/go-openapi/errors v0.22.9 // indirect
	github.com/go-openapi/jsonpointer v1.0.1 // indirect
	github.com/go-openapi/jsonreference v1.0.2 // indirect
	github.com/go-openapi/loads v0.25.3 // indirect
	github.com/go-openapi/runtime v0.33.3 // indirect
	github.com/go-openapi/runtime/server-middleware v0.33.3 // indirect
	github.com/go-openapi/spec v1.0.1 // indirect
	github.com/go-openapi/strfmt v0.27.4 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-openapi/validate v1.0.0 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/cel-go v0.31.0 // indirect
	github.com/google/certificate-transparency-go v1.4.0 // indirect
	github.com/google/go-containerregistry v0.22.1 // indirect
	github.com/google/go-github/v88 v88.0.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/gosimple/unidecode v1.0.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.32.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.8 // indirect
	github.com/in-toto/in-toto-golang v0.11.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jedisct1/go-minisign v0.0.0-20211028175153-1c139d1cc84b // indirect
	github.com/kevinburke/ssh_config v1.6.0 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mailru/easyjson v0.9.2 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/oklog/ulid/v2 v2.1.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/otiai10/mint v1.6.3 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/philhofer/fwd v1.2.1 // indirect
	github.com/pjbgf/sha1cd v0.7.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/sabhiram/go-gitignore v0.0.0-20210923224102-525f6e181f06 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/samber/slog-common v0.22.0 // indirect
	github.com/samber/slog-zerolog/v2 v2.9.2 // indirect
	github.com/sassoftware/relic/v8 v8.2.0 // indirect
	github.com/secure-systems-lab/go-securesystemslib v0.11.1 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/shibumi/go-pathspec v1.3.0 // indirect
	github.com/sigstore/protobuf-specs v0.5.2 // indirect
	github.com/sigstore/rekor v1.5.5 // indirect
	github.com/sigstore/rekor-tiles/v2 v2.4.0 // indirect
	github.com/sigstore/sigstore v1.11.1 // indirect
	github.com/sigstore/timestamp-authority/v2 v2.2.0 // indirect
	github.com/skeema/knownhosts v1.3.3 // indirect
	github.com/spf13/pflag v1.1.0 // indirect
	github.com/stretchr/objx v0.8.0 // indirect
	github.com/theupdateframework/go-tuf v0.7.0 // indirect
	github.com/theupdateframework/go-tuf/v2 v2.4.2 // indirect
	github.com/tinylib/msgp v1.6.5 // indirect
	github.com/transparency-dev/formats v0.1.1 // indirect
	github.com/transparency-dev/merkle v0.0.2 // indirect
	github.com/xanzy/ssh-agent v0.3.3 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/crypto v0.59.0 // indirect
	golang.org/x/exp v0.0.0-20261113113407-e3970619a1a4 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/term v0.48.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261110000943-09e21b237f2c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261110000943-09e21b237f2c // indirect
	google.golang.org/grpc v1.84.0 // indirect
	gopkg.in/ini.v1 v1.68.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
)

exclude (
	github.com/sergi/go-diff v1.2.0
	github.com/sergi/go-diff v1.3.0
	github.com/sergi/go-diff v1.3.1
)
