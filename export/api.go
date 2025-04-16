package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

const (
	dataVolumePath = "/data"
)

type Provider interface {
	GetExportDogu(ctx context.Context) (*doguExport, error)
	SetExportDogu(doguName string, ctx context.Context) (*doguExport, error)
	GetExportMode(ctx context.Context) (*exportModeStatus, error)
}

type Controller struct {
	provider Provider
}

func NewController(provider Provider) *Controller {
	return &Controller{provider: provider}
}

func (m *Controller) GetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguExp, err := m.provider.GetExportDogu(r.Context())
	if err != nil {
		core.ErrorResponse(w, http.StatusNotFound, fmt.Sprintf("failed to get export dogu: %s", err.Error()))
		return
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *Controller) SetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	doguExp, err := m.provider.SetExportDogu(doguName, r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to set export dogu: %w", err))
		return
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *Controller) GetExportMode(w http.ResponseWriter, r *http.Request) {
	// delegate call to multinode or classic provider
	mode, err := m.provider.GetExportMode(r.Context())

	// internal error if provider fails
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, mode)
}
