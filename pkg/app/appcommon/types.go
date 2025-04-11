package appcommon

import (
	"github.com/cidverse/go-vcsapp/pkg/platform/api"
)

type VCSEnvironment struct {
	Env  api.CIEnvironment
	Vars []api.CIVariable
}
