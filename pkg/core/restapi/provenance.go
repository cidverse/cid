package restapi

import (
	"net/http"

	"github.com/cidverse/cid/pkg/core/provenance"
	"github.com/labstack/echo/v4"
)

// projectInformation returns all available information about the current project
func (hc *handlerConfig) provenance(c echo.Context) error {
	prov := provenance.GenerateProvenance(hc.env, hc.state)

	return c.JSON(http.StatusOK, prov)
}
