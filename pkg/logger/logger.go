// pkg/logger/logger.go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

// Initialize sets up our logger
func Initialize(debug bool) {
    var cfg zap.Config
    if debug {
        cfg = zap.NewDevelopmentConfig()
        cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    } else {
        cfg = zap.NewProductionConfig()
    }

    baseLogger, _ := cfg.Build()
    sugar = baseLogger.Sugar()
}

// Get returns the sugared logger
func Get() *zap.SugaredLogger {
    if sugar == nil {
        Initialize(true) // Default to debug mode if not initialized
    }
    return sugar
}

// Sync flushes any buffered log entries
func Sync() {
    if sugar != nil {
        sugar.Sync()
    }
}