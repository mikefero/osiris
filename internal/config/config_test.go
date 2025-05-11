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
package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mikefero/osiris/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("verify defaults are set when overrides are not provided", func(t *testing.T) {
		actual, err := config.NewConfig()
		require.NoError(t, err)

		expected := &config.Config{
			BaseURL:        "http://localhost:3737",
			ControlPlaneID: uuid.MustParse("4168295f-015e-4190-837e-0fcc5d72a52f"),
			Logger: config.Logger{
				Level:     "info",
				Filename:  "osiris.log",
				Retention: 7,
			},
			OutputFile: "osiris.json",
			Sanitize:   true,
			Timeouts: config.Timeouts{
				Timeout:        15 * time.Second,
				ResponseHeader: 15 * time.Second,
			},
		}
		require.Equal(t, expected, actual)
	})

	t.Run("verify overrides are set when overrides are provided", func(t *testing.T) {
		t.Setenv("OSIRIS_BASE_URL", "http://example.com")
		t.Setenv("OSIRIS_BEARER_TOKEN", "test-token-123")
		t.Setenv("OSIRIS_CONTROL_PLANE_ID", "37b0c1f3-4a2e-4d5b-8f7c-9a2e6d5f3a1b")
		t.Setenv("OSIRIS_LOGGER_LEVEL", "debug")
		t.Setenv("OSIRIS_LOGGER_FILENAME", "osiris-debug.log")
		t.Setenv("OSIRIS_LOGGER_RETENTION", "14")
		t.Setenv("OSIRIS_OUTPUT_FILE", "output.json")
		t.Setenv("OSIRIS_SANITIZE", "false")
		t.Setenv("OSIRIS_TIMEOUTS_TIMEOUT", "20s")
		t.Setenv("OSIRIS_TIMEOUTS_RESPONSE_HEADER", "25s")
		actual, err := config.NewConfig()
		require.NoError(t, err)

		expected := &config.Config{
			BaseURL:        "http://example.com",
			BearerToken:    "test-token-123",
			ControlPlaneID: uuid.MustParse("37b0c1f3-4a2e-4d5b-8f7c-9a2e6d5f3a1b"),
			Logger: config.Logger{
				Level:     "debug",
				Filename:  "osiris-debug.log",
				Retention: 14,
			},
			OutputFile: "output.json",
			Sanitize:   false,
			Timeouts: config.Timeouts{
				Timeout:        20 * time.Second,
				ResponseHeader: 25 * time.Second,
			},
		}
		require.Equal(t, expected, actual)
	})

	t.Run("verify invalid UUID in environment variable returns error", func(t *testing.T) {
		t.Setenv("OSIRIS_CONTROL_PLANE_ID", "not-a-valid-uuid")
		_, err := config.NewConfig()
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid UUID")
	})

	t.Run("verify bearer token is properly set from environment", func(t *testing.T) {
		t.Setenv("OSIRIS_BEARER_TOKEN", "test-token-123")
		actual, err := config.NewConfig()
		require.NoError(t, err)
		require.Equal(t, "test-token-123", actual.BearerToken)
	})

	t.Run("verify partial overrides work correctly", func(t *testing.T) {
		// Only override some settings, not all
		t.Setenv("OSIRIS_BASE_URL", "http://partial-example.com")
		t.Setenv("OSIRIS_TIMEOUTS_TIMEOUT", "45s")

		actual, err := config.NewConfig()
		require.NoError(t, err)

		// Should have our overridden values
		require.Equal(t, "http://partial-example.com", actual.BaseURL)
		require.Equal(t, 45*time.Second, actual.Timeouts.Timeout)

		// Other values should still be defaults
		require.Equal(t, uuid.MustParse("4168295f-015e-4190-837e-0fcc5d72a52f"), actual.ControlPlaneID)
		require.True(t, actual.Sanitize)
		require.Equal(t, "osiris.json", actual.OutputFile)
		require.Equal(t, 15*time.Second, actual.Timeouts.ResponseHeader)
	})

	t.Run("verify configuration is loaded from configuration file", func(t *testing.T) {
		dir := t.TempDir()
		file, err := os.Create(filepath.Join(dir, "osiris.yaml"))
		if err != nil {
			t.Fatalf("unable to create config file: %v", err)
		}
		_, err = file.Write([]byte(`base_url: http://example.com
bearer_token: test-token-123
control_plane_id: 37b0c1f3-4a2e-4d5b-8f7c-9a2e6d5f3a1b
logger:
  level: debug
  filename: osiris-debug.log
  retention: 14
output_file: output.json
sanitize: false
timeouts:
  timeout: 20s
  response_header: 25s
`))
		if err != nil {
			t.Fatalf("unable to write config file: %v", err)
		}
		require.NoError(t, file.Close())
		viper.AddConfigPath(dir)
		defer viper.Reset()
		actual, err := config.NewConfig()
		require.NoError(t, err)

		expected := &config.Config{
			BaseURL:        "http://example.com",
			BearerToken:    "test-token-123",
			ControlPlaneID: uuid.MustParse("37b0c1f3-4a2e-4d5b-8f7c-9a2e6d5f3a1b"),
			Logger: config.Logger{
				Level:     "debug",
				Filename:  "osiris-debug.log",
				Retention: 14,
			},
			OutputFile: "output.json",
			Sanitize:   false,
			Timeouts: config.Timeouts{
				Timeout:        20 * time.Second,
				ResponseHeader: 25 * time.Second,
			},
		}
		require.Equal(t, expected, actual)
	})

	t.Run("verify environment variables take precedence over config file", func(t *testing.T) {
		dir := t.TempDir()
		file, err := os.Create(filepath.Join(dir, "osiris.yaml"))
		if err != nil {
			t.Fatalf("unable to create config file: %v", err)
		}
		_, err = file.Write([]byte(`base_url: http://example.com
bearer_token: test-token-123
control_plane_id: 37b0c1f3-4a2e-4d5b-8f7c-9a2e6d5f3a1b
logger:
  level: debug
  filename: osiris-debug.log
  retention: 14
output_file: output.json
sanitize: false
timeouts:
  timeout: 20s
  response_header: 25s
`))
		if err != nil {
			t.Fatalf("unable to write config file: %v", err)
		}
		require.NoError(t, file.Close())
		viper.AddConfigPath(dir)
		defer viper.Reset()

		// Override with environment variables
		t.Setenv("OSIRIS_BASE_URL", "http://environment.com")
		t.Setenv("OSIRIS_BEARER_TOKEN", "environment-test-token-123")
		t.Setenv("OSIRIS_CONTROL_PLANE_ID", "869b5090-71bd-4387-be27-567d67ec286d")

		actual, err := config.NewConfig()
		require.NoError(t, err)

		// Environment variables should take precedence; other values should come
		// from config file
		expected := &config.Config{
			BaseURL:        "http://environment.com",
			BearerToken:    "environment-test-token-123",
			ControlPlaneID: uuid.MustParse("869b5090-71bd-4387-be27-567d67ec286d"),
			Logger: config.Logger{
				Level:     "debug",
				Filename:  "osiris-debug.log",
				Retention: 14,
			},
			OutputFile: "output.json",
			Sanitize:   false,
			Timeouts: config.Timeouts{
				Timeout:        20 * time.Second,
				ResponseHeader: 25 * time.Second,
			},
		}
		require.Equal(t, expected, actual)
	})

	t.Run("verify invalid time duration returns error", func(t *testing.T) {
		t.Setenv("OSIRIS_TIMEOUTS_TIMEOUT", "not-a-valid-duration")
		_, err := config.NewConfig()
		require.Error(t, err)
		require.Contains(t, err.Error(), "time: invalid duration")
	})
}
