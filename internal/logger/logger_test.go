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
package logger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikefero/osiris/internal/config"
	"github.com/mikefero/osiris/internal/logger"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	t.Run("verify logger commands", func(t *testing.T) {
		tests := []struct {
			name     string
			cmdType  logger.LoggerCommandType
			expected string
		}{
			{
				name:     "dump command",
				cmdType:  logger.LoggerCommandTypeDump,
				expected: "dump",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.expected, tt.cmdType.String())
			})
		}
	})

	t.Run("verify logger levels and filenames", func(t *testing.T) {
		tests := []struct {
			name        string
			level       string
			expected    zapcore.Level
			expectError bool
		}{
			{
				name:        "debug level",
				level:       "debug",
				expected:    zap.DebugLevel,
				expectError: false,
			},
			{
				name:        "info level",
				level:       "info",
				expected:    zap.InfoLevel,
				expectError: false,
			},
			{
				name:        "warn level",
				level:       "warn",
				expected:    zap.WarnLevel,
				expectError: false,
			},
			{
				name:        "error level",
				level:       "error",
				expected:    zap.ErrorLevel,
				expectError: false,
			},
			{
				name:        "panic level",
				level:       "panic",
				expected:    zap.PanicLevel,
				expectError: false,
			},
			{
				name:        "fatal level",
				level:       "fatal",
				expected:    zap.FatalLevel,
				expectError: false,
			},
			{
				name:        "invalid level",
				level:       "invalid",
				expectError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := t.TempDir()
				config := config.Logger{
					Level:    tt.level,
					Filename: filepath.Join(dir, tt.level) + ".log",
				}
				logger, err := logger.NewLogger(config, logger.LoggerCommandTypeDump)

				if tt.expectError {
					require.Error(t, err)
					require.Nil(t, logger)
				} else {
					require.NoError(t, err)
					require.NotNil(t, logger)
					require.Equal(t, tt.expected, logger.Level())

					// Fatal level should not be called as zap.Logger exits the process
					// Panic level should not be called as zap.Logger creates a panic
					if tt.level != "panic" && tt.level != "fatal" {
						logger.Error("test message")
						require.NoError(t, logger.Sync())
						_, err := os.Stat(config.Filename)
						require.NoError(t, err)
					}
				}
			})
		}
	})
}
