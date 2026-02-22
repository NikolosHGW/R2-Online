// Package logger configures a zap.Logger for the application.
//
// Log level is controlled by the LOG_LEVEL env var:
//
//	LOG_LEVEL=debug   → все пакеты, hex-дампы, трассировка
//	LOG_LEVEL=info    → подключения, аутентификация, ошибки     (default)
//	LOG_LEVEL=warn    → только предупреждения и ошибки
//	LOG_LEVEL=error   → только ошибки
//
// In production set LOG_LEVEL=info (or omit it).
// For debugging packet issues set LOG_LEVEL=debug.
package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds a zap.Logger from the LOG_LEVEL environment variable.
func New() *zap.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "time"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // colorised in terminals

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		level,
	)

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(0))
}

func parseLevel(s string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zap.DebugLevel
	case "warn", "warning":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	default:
		return zap.InfoLevel
	}
}
