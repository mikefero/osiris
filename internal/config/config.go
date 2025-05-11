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
package config

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

const (
	defaultBaseURL               = "http://localhost:3737"
	defaultSanitize              = true
	defaultOutputFile            = "osiris.json"
	defaultTimeoutTimeout        = 15 * time.Second
	defaultTimeoutResponseHeader = 15 * time.Second
)

var defaultControlPlaneID = uuid.MustParse("4168295f-015e-4190-837e-0fcc5d72a52f")

// Config is the configuration struct for osiris.
// It contains the base URL for the admin API, the bearer token for
// authenticating with the admin API, the control plane ID for the
// GET/PUT/POST requests, the logger configuration, and the timeouts for
// the API requests.
type Config struct {
	// BaseURL is the base URL for the admin API.
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
	// BearerToken is the bearer token for authenticating with the admin API.
	BearerToken string `yaml:"bearer_token" mapstructure:"bearer_token"`
	// ControlPlaneID is the control plane ID for the GET/PUT/POST requests.
	ControlPlaneID uuid.UUID `yaml:"control_plane_id" mapstructure:"control_plane_id"`
	// Logger is the logger configuration.
	Logger Logger `yaml:"logger" mapstructure:"logger"`
	// Sanitize is a flag to enable or disable sanitization of the response body
	// fields.
	Sanitize bool `yaml:"sanitize" mapstructure:"sanitize"`
	// OutputFile is the output file for the sanitized configuration of a control
	// plane.
	OutputFile string `yaml:"output_file" mapstructure:"output_file"`
	// Timeouts are the timeouts for the API requests.
	Timeouts Timeouts `yaml:"timeouts" mapstructure:"timeouts"`
}

// Logger is the logger configuration for osiris.
// It contains the log level, the log file name, and the number of days to
// retain the log files.
type Logger struct {
	// Level is the log level for the logger.
	Level string `yaml:"level" mapstructure:"level"`
	// Filename is the log file name for the logger.
	Filename string `yaml:"filename" mapstructure:"filename"`
	// Retention is the number of days to retain the log files.
	Retention int `yaml:"retention" mapstructure:"retention"`
}

// Timeouts is the timeouts configuration for osiris.
type Timeouts struct {
	// Timeout is the timeout for request by the client.
	Timeout time.Duration `yaml:"timeout" mapstructure:"timeout"`
	// ResponseHeader is the timeout for reading the headers.
	ResponseHeader time.Duration `yaml:"response_header" mapstructure:"response_header"`
}

func NewConfig() (*Config, error) {
	// Defaults
	viper.SetDefault("base_url", defaultBaseURL)
	viper.SetDefault("control_plane_id", defaultControlPlaneID)
	viper.SetDefault("output_file", defaultOutputFile)
	viper.SetDefault("sanitize", defaultSanitize)

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.filename", "osiris.log")
	viper.SetDefault("logger.retention", 7)

	// Timeout defaults
	viper.SetDefault("timeouts.timeout", defaultTimeoutTimeout)
	viper.SetDefault("timeouts.response_header", defaultTimeoutResponseHeader)

	// Osiris configuration setup for viper
	viper.SetConfigName("osiris")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables to viper that do not have a corresponding
	// default value
	viper.SetEnvPrefix("osiris")
	if err := viper.BindEnv("bearer_token"); err != nil {
		return nil, fmt.Errorf("unable to bind bearer_token environment variable: %w", err)
	}

	// Enable automatic environment variable binding
	viper.AutomaticEnv()

	// Read in the configuration file and ignore not found errors as environment
	// variables will be used if the file is not found. If the required
	// configuration fields are not present then and error will be returned
	// further down the line.
	var config Config
	_ = viper.ReadInConfig()
	err := viper.Unmarshal(&config, viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			// Custom UUID conversion hook
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if f.Kind() != reflect.String {
					return data, nil
				}

				if t != reflect.TypeOf(uuid.UUID{}) {
					return data, nil
				}

				strData, ok := data.(string)
				if !ok {
					return nil, fmt.Errorf("failed type assertion to string")
				}

				return uuid.Parse(strData)
			},

			// Use built-in time.Duration decoder
			mapstructure.StringToTimeDurationHookFunc(),
		),
	))
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal config: %w", err)
	}
	return &config, nil
}
