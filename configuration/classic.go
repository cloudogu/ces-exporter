package configuration

import (
	"context"
	"log/slog"
)

var _ Provider = (*ClassicConfigurationProvider)(nil)

type ClassicConfigurationProvider struct {
}

func NewClassicConfigurationProvider() *ClassicConfigurationProvider {
	return &ClassicConfigurationProvider{}
}

//func StartCLI() error {
//	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
//		Level: slog.LevelInfo, // Set log level to Debug
//	}))
//
//	// Set the logger as the default one
//	slog.SetDefault(logger)
//
//	doguStrList := flag.String("d", "", "Comma-separated list of dogus to export config (default: all)")
//	ignoreKeys := flag.String("i", "", "Comma-separated list of keys or subkeys to ignore (default: none)")
//	outputFile := flag.String("o", "", "Output file path (default: stdout)")
//	help := flag.Bool("h", false, "Show usage help")
//
//	flag.Parse()
//
//	if *help {
//		printHelp()
//		return nil
//	}
//
//	keyIgnoreList := extractFromString(ignoreKeys)
//	slog.Info("ignoring keys", "keyIgnoreList", keyIgnoreList)
//	doguList := extractFromString(doguStrList)
//
//	if len(doguList) == 0 {
//		allDogus, err := GetAllDogus()
//		if err != nil {
//			return fmt.Errorf("failed to get all dogus: %v", err)
//		}
//
//		doguList = allDogus
//	}
//
//	slog.Info("Created dogu list for config export", "dogus", doguList)
//
//	export, err := ExportConfigs(doguList, keyIgnoreList)
//	if err != nil {
//		return err
//	}
//
//	slog.Info("Finished getting config from system, try to write config...")
//
//	if *outputFile == "" {
//		fmt.Println(export)
//		return nil
//	}
//
//	if lErr := writeToFile(outputFile, export); lErr != nil {
//		return fmt.Errorf("failed to write to file: %v", lErr)
//	}
//
//	slog.Info("Export config completed")
//
//	return nil
//}

func (c ClassicConfigurationProvider) getBackupSchedules(ctx context.Context) ([]backupSchedule, error) {
	return []backupSchedule{}, nil
}

func (c ClassicConfigurationProvider) getGlobalConfigs(ctx context.Context) ([]keyValue, error) {
	return nil, nil
}

func (c ClassicConfigurationProvider) getDoguConfigs(ctx context.Context) ([]doguConfig, error) {
	slog.Debug("get dogu configs...")
	return nil, nil
}
