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
package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mikefero/osiris/internal/config"
	"go.uber.org/zap"
)

const defaultRateLimitWaitDuration = 10 * time.Second

// HTTPClient is an interface that wraps the Do method of http.Client.
type HTTPClient interface {
	// Do executes a single HTTP request and returns the response or an error
	// if the request fails.
	Do(req *http.Request) (*http.Response, error)
}

// Client is a struct that represents the API client.
type Client struct {
	httpClient     HTTPClient
	baseURL        string
	bearerToken    string
	outputFilename string
	logger         *zap.Logger
}

// NewClient creates a new API client with the provided configuration and logger.
func NewClient(config *config.Config, logger *zap.Logger) *Client {
	client := &http.Client{
		Timeout: config.Timeouts.Timeout,
		Transport: &http.Transport{
			ResponseHeaderTimeout: config.Timeouts.ResponseHeader,
		},
	}
	baseURL := fmt.Sprintf("%s/%s", strings.TrimSuffix(config.BaseURL, "/"),
		config.ControlPlaneID.String())

	return &Client{
		httpClient:     client,
		baseURL:        baseURL,
		bearerToken:    config.BearerToken,
		outputFilename: config.OutputFile,
		logger: logger.With(
			zap.String("base-url", baseURL),
			zap.Any("control-plane-id", config.ControlPlaneID),
		),
	}
}

// FetchData implements the resource.APIClient interface method to fetch data
// from an endpoint.
func (c *Client) FetchData(ctx context.Context, endpoint string) ([]map[string]interface{}, error) {
	return c.fetchEndpoint(ctx, endpoint)
}
