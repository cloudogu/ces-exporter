package export

import (
	"context"
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
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	//TODO implement me

	doguExp := &DoguExport{
		Dogu: doguName,
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *MultinodeExportModeController) SetExportDogu(w http.ResponseWriter, r *http.Request) {
	doguName := r.PathValue("doguName")
	if doguName == "" {
		core.BadRequest(w, "doguName must not be empty")
		return
	}

	//TODO implement me

	doguExp := &DoguExport{
		Dogu: doguName,
	}

	core.JSON(w, http.StatusOK, doguExp)
}

func (m *MultinodeExportModeController) GetExportMode(w http.ResponseWriter, r *http.Request) {
	status := &ExportModeStatus{
		IsActive: false,
	}

	//TODO implement me

	core.JSON(w, http.StatusOK, status)
}
