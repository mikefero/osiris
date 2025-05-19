/*
Copyright © 2025 Michael Fero

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package logger

import (
	"fmt"

	"github.com/mikefero/osiris/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LoggerCommandType is the type of command for the logger.
type LoggerCommandType int

const (
	// LoggerCommandTypeDump is the command type for dump.
	LoggerCommandTypeDump LoggerCommandType = iota
	// LoggerCommandTypeReset is the command type for reset.
	LoggerCommandTypeReset
)

// LoggerCommandTypeString returns the string representation of the command type.
func (l LoggerCommandType) String() string {
	return [...]string{
		"dump",
		"reset",
	}[l]
}

// NewLogger creates a new zap logger with the specified configuration and command type.
// It uses lumberjack for log rotation and compression.
// The log level is set based on the configuration.
// The command type is added as a field to the logger.
// Returns a zap.Logger instance and an error if any occurs during creation.
func NewLogger(config config.Logger, commandType LoggerCommandType) (*zap.Logger, error) {
	zapLoggerLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("unable to parse log level: %w", err)
	}

	// Add daily log rotator for zap logger
	logger := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    0, // unlimited
		MaxBackups: config.Retention,
		MaxAge:     config.Retention,
		Compress:   true,
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logger),
		zapLoggerLevel,
	).With([]zapcore.Field{
		zap.String("command", commandType.String()),
	})
	zapLogger := zap.New(core)
	return zapLogger, nil
}
