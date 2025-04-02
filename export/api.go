package export

import (
	"context"
	"fmt"
	"github.com/cloudogu/ces-exporter/core"
	"net/http"
)

type Provider interface {
	GetExportDogu(ctx context.Context) (*DoguExport, error)
	SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error)
	GetExportMode(ctx context.Context) (*ExportModeStatus, error)
}

type MultinodeExportModeController struct {
	provider Provider
}

func NewMultinodeExportModeController(provider Provider) *MultinodeExportModeController {
	return &MultinodeExportModeController{provider: provider}
}

func (m *MultinodeExportModeController) GetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguExp, err := m.provider.GetExportDogu(r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to get export dogu: %w", err))
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *MultinodeExportModeController) SetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	doguExp, err := m.provider.SetExportDogu(doguName, r.Context())
	if err != nil {
		core.InternalServerErrorResponse(w, fmt.Errorf("failed to set export dogu: %w", err))
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *MultinodeExportModeController) GetExportMode(w http.ResponseWriter, r *http.Request) {
	// delegate call to multinode or classic provider
	mode, err := m.provider.GetExportMode(r.Context())

	// internal error if provider fails
	if err != nil {
		core.InternalServerErrorResponse(w, err)
		return
	}

	core.JSON(w, http.StatusOK, mode)
}
