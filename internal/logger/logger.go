package logger

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	ProductionEnv = "prod"
	LocalEnv      = "local"
)

var Log *zap.SugaredLogger = zap.NewNop().Sugar()

func Initialize(level string, env string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	var cfg zap.Config
	switch env {
	case ProductionEnv:
		cfg = zap.NewProductionConfig()
	case LocalEnv:
		cfg = zap.NewDevelopmentConfig()
	default:
		return fmt.Errorf("logger valid environment values: %s, %s", ProductionEnv, LocalEnv)
	}

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl.Sugar()
	return nil
}
