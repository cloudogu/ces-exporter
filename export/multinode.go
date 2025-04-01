package export

import "context"

type MultinodeExportModeProvider struct{}

func NewMultinodeExportModeProvider() *MultinodeExportModeProvider {
	return &MultinodeExportModeProvider{}
}

func (m MultinodeExportModeProvider) GetExportDogu(ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) SetExportDogu(doguName string, ctx context.Context) (*DoguExport, error) {
	return nil, nil
}

func (m MultinodeExportModeProvider) GetExportMode(ctx context.Context) (*ExportModeStatus, error) {
	return nil, nil
}
