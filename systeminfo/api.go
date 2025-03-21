package systeminfo

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

const (
	globalConfigKeyFqdn = "fqdn"
)

type Controller struct {
	systemInfoProvider SystemInfoProvider
}

type SystemInfoProvider interface {
	getComponents(ctx context.Context) ([]component, error)
	getDogus(ctx context.Context) ([]dogu, error)
	getFqdn(ctx context.Context) (string, error)
	isMultinode() bool
}

func NewController(provider SystemInfoProvider) *Controller {
	return &Controller{
		systemInfoProvider: provider,
	}
}

func (sic Controller) GetSystemInfo(w http.ResponseWriter, r *http.Request) {
	fqdn, err := sic.systemInfoProvider.getFqdn(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get fqdn: %w", err))
		return
	}

	dogus, err := sic.systemInfoProvider.getDogus(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get dogus: %w", err))
		return
	}

	components, err := sic.systemInfoProvider.getComponents(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get components: %w", err))
		return
	}

	info := &systemInfo{
		FQDN:        fqdn,
		IsMultinode: sic.systemInfoProvider.isMultinode(),
		Dogus:       dogus,
		Components:  components,
	}

	core.JSON(w, http.StatusOK, info)
}
