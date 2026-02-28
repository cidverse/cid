package restapi

import (
	"github.com/cidverse/cid/pkg/core/actionsdk"
)

type APIConfig struct {
	SDKClient actionsdk.SDKClient
}

// apiError, see https://www.rfc-editor.org/rfc/rfc7807
type apiError struct {
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}
