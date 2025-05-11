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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mikefero/osiris/internal/resource"
	"go.uber.org/zap"
)

// GatherData fetches data from all resources in the resource registry.
// It uses goroutines to fetch data concurrently and collects the results.
func (c *Client) GatherData(ctx context.Context) error {
	resources := resource.ResourceRegistry
	errChan := make(chan error, len(resources))
	var mutex sync.Mutex
	var results []resource.ResourceData
	var wg sync.WaitGroup

	c.logger.Info("Gathering data from resources",
		zap.Int("resource_count", len(resources)))

	// Iterate over the resources and start a goroutine for each one
	startTime := time.Now()
	for _, res := range resources {
		wg.Add(1)
		go func(res resource.Resource) {
			defer wg.Done()

			// List the resource items
			data, err := res.List(ctx, c)
			if err != nil {
				c.logger.Error("error listing resource",
					zap.String("resource", res.Name()),
					zap.Error(err))
				errChan <- fmt.Errorf("error listing resource %s: %w", res.Name(), err)
				return
			}

			if len(data) > 0 {
				c.logger.Info("Fetched data for resource",
					zap.String("resource", res.Name()),
					zap.Int("items", len(data)),
					zap.Duration("fetch-duration", time.Since(startTime)))

				mutex.Lock()
				results = append(results, resource.ResourceData{
					Data: data,
					Name: res.Name(),
				})
				mutex.Unlock()
			} else {
				c.logger.Debug("No data found for resource",
					zap.String("resource", res.Name()))
			}
		}(res)
	}

	// Rest of the function remains the same...
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		c.logger.Warn("Context was canceled while fetching data from resources",
			zap.Error(ctx.Err()))
		return ctx.Err()
	case <-done:
		close(errChan)
		if len(errChan) > 0 {
			err := <-errChan
			c.logger.Error("Error occurred while fetching data from resources",
				zap.Error(err))
			return err
		}
	}

	c.logger.Info("Successfully gathered data from resources",
		zap.Int("resource-count", len(resources)),
		zap.Duration("duration", time.Since(startTime)))

	if err := c.writeResults(results); err != nil {
		c.logger.Error("error writing results",
			zap.String("output-filename", c.outputFilename),
			zap.Error(err))
		return fmt.Errorf("error writing results: %w", err)
	}
	return nil
}

func (c *Client) fetchEndpoint(ctx context.Context, endpoint string) ([]map[string]interface{}, error) {
	endpointURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	var result []map[string]interface{}

	c.logger.Debug("Fetching endpoint",
		zap.String("endpoint", endpoint),
		zap.String("endpoint-url", endpointURL))

	pageCount := 0
	pageURL := endpointURL
	startTime := time.Now()
	for len(pageURL) > 0 {
		requestStartTime := time.Now()
		if err := ctx.Err(); err != nil {
			c.logger.Warn("Context canceled during pagination",
				zap.String("endpoint", endpoint),
				zap.String("endpoint-url", endpointURL),
				zap.Error(err))
			return nil, err
		}

		pageCount++
		c.logger.Debug("Fetching page",
			zap.String("endpoint", endpoint),
			zap.String("page-url", pageURL),
			zap.Int("page-number", pageCount))

		data, nextPageURL, err := c.fetchEndpointPage(ctx, pageURL)
		if err != nil {
			// Check if the error is a RateLimitError
			errRateLimit, ok := err.(*RateLimitError)
			if !ok {
				return nil, fmt.Errorf("error fetching endpoint %s: %w", endpoint, err)
			}

			// Handle rate limit Retry-After duration
			c.logger.Warn("Rate limit exceeded; retrying",
				zap.String("endpoint", endpoint),
				zap.String("page-url", pageURL),
				zap.Int("page-number", pageCount),
				zap.Duration("retry-after", errRateLimit.RetryAfter),
				zap.Duration("request-duration", time.Since(requestStartTime)))

			time.Sleep(errRateLimit.RetryAfter)
			continue
		}

		if len(data) == 0 {
			c.logger.Debug("No data found for endpoint",
				zap.String("endpoint-url", pageURL),
				zap.Duration("request-duration", time.Since(requestStartTime)))
			return nil, nil
		}

		c.logger.Debug("Fetched data from page",
			zap.String("endpoint", endpoint),
			zap.String("page-url", pageURL),
			zap.Int("page-number", pageCount),
			zap.Int("item-count", len(data)),
			zap.Duration("request-duration", time.Since(requestStartTime)))

		result = append(result, data...)

		if len(nextPageURL) == 0 {
			c.logger.Debug("No more pages to fetch",
				zap.String("endpoint", endpoint),
				zap.String("page-url", pageURL))
			break
		}
		pageURL = nextPageURL
	}

	c.logger.Debug("Fetched all pages",
		zap.String("endpoint", endpoint),
		zap.Int("total-pages", pageCount),
		zap.Int("total-items", len(result)),
		zap.Duration("fetch-duration", time.Since(startTime)))

	return result, nil
}

func (c *Client) fetchEndpointPage(ctx context.Context, url string) ([]map[string]interface{}, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %w", err)
	}

	// Set the Authorization header with the bearer token and execute the request
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.bearerToken))
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("error making request",
			zap.String("url", url),
			zap.Duration("request-duration", time.Since(startTime)),
			zap.Error(err))
		return nil, "", fmt.Errorf("error making request: %w", err)
	}
	//nolint: errcheck
	defer resp.Body.Close()

	c.logger.Debug("Received response",
		zap.String("url", url),
		zap.Int("status", resp.StatusCode),
		zap.Duration("request-duration", time.Since(startTime)))

	// Check the status code and handle the response accordingly
	startTime = time.Now()
	switch resp.StatusCode {
	case http.StatusOK:
		var pageResp = struct {
			Data []map[string]interface{} `json:"data"`
			Next string                   `json:"next"`

			Items []map[string]interface{} `json:"items"`
			Page  struct {
				HasNextPage bool   `json:"has_next_page"`
				TotalCount  int    `json:"total_count"`
				NextCursor  string `json:"next_cursor"`
			} `json:"page"`
		}{}
		if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
			c.logger.Error("error decoding response",
				zap.String("url", url),
				zap.Error(err))
			return nil, "", fmt.Errorf("error decoding response: %w", err)
		}

		// Remove unwanted fields from each item
		if len(pageResp.Data) > 0 {
			for _, item := range pageResp.Data {
				delete(item, "updated_at")
				delete(item, "created_at")
			}
		} else if len(pageResp.Items) > 0 {
			// Handle v1 API response
			for _, item := range pageResp.Items {
				delete(item, "updated_at")
				delete(item, "created_at")
			}
			pageResp.Data = pageResp.Items
		}

		c.logger.Debug("Parsed response",
			zap.String("url", url),
			zap.String("next", pageResp.Next),
			zap.Int("item-count", len(pageResp.Data)),
			zap.Duration("parse-duration", time.Since(startTime)))

		// Determine the next URL to fetch
		var nextURL string
		if len(pageResp.Next) > 0 {
			nextURL = fmt.Sprintf("%s/%s", c.baseURL, strings.TrimPrefix(pageResp.Next, "/"))
			c.logger.Debug("Next URL found",
				zap.String("url", url),
				zap.String("next-url", nextURL))
		} else if pageResp.Page.HasNextPage {
			// Handle v1 API pagination with cursor
			nextURL = fmt.Sprintf("%s?page.next_cursor=%s", url, pageResp.Page.NextCursor)
			c.logger.Debug("Next URL found with cursor",
				zap.String("url", url),
				zap.String("next-url", nextURL))
		}

		return pageResp.Data, nextURL, nil
	case http.StatusTooManyRequests:
		retryDuration := c.getRetryAfterDuration(resp)
		c.logger.Warn("Rate limit exceeded; retrying",
			zap.String("url", url),
			zap.Duration("retry-after", retryDuration))
		return nil, url, &RateLimitError{RetryAfter: retryDuration}
	case http.StatusNotFound:
		c.logger.Error("Endpoint not found",
			zap.String("url", url),
			zap.Int("status-code", resp.StatusCode))
		return nil, "", nil
	default:
		c.logger.Error("unhandled status code",
			zap.String("url", url),
			zap.Int("status-code", resp.StatusCode))
		return nil, "", fmt.Errorf("unhandled status code: %d", resp.StatusCode)
	}
}

func (c *Client) getRetryAfterDuration(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if len(retryAfter) == 0 {
		c.logger.Debug("Retry-After header not found; using default duration",
			zap.Duration("duration", defaultRateLimitWaitDuration))
		return defaultRateLimitWaitDuration
	}

	duration, err := time.ParseDuration(retryAfter)
	if err != nil {
		c.logger.Error("error parsing Retry-After header; using default duration",
			zap.Duration("duration", defaultRateLimitWaitDuration),
			zap.String("retry-after", retryAfter),
			zap.Error(err))
		return defaultRateLimitWaitDuration
	}
	return duration
}

func (c *Client) writeResults(results []resource.ResourceData) error {
	// Create a map where the keys are the endpoint names
	resultMap := make(map[string][]map[string]interface{})

	// Convert the slice of Results to a map
	for _, result := range results {
		resultMap[result.Name] = result.Data
	}

	c.logger.Info("Marshaling results to JSON",
		zap.Int("endpointCount", len(resultMap)))

	// Marshal the map to JSON with pretty formatting
	startTime := time.Now()
	jsonData, err := json.MarshalIndent(resultMap, "", "  ")
	if err != nil {
		c.logger.Error("error marshaling results", zap.Error(err))
		return fmt.Errorf("error marshaling results: %w", err)
	}

	c.logger.Debug("Writing results to file",
		zap.String("output-filename", c.outputFilename),
		zap.Int("bytes", len(jsonData)),
		zap.Duration("duration", time.Since(startTime)))

	if err := os.WriteFile(c.outputFilename, jsonData, 0600); err != nil {
		c.logger.Error("error writing file",
			zap.String("outout-filename", c.outputFilename),
			zap.Error(err))
		return fmt.Errorf("error writing file: %w", err)
	}

	c.logger.Info("Successfully wrote results to JSON file",
		zap.String("output-filename", c.outputFilename),
		zap.Int("bytes", len(jsonData)),
		zap.Duration("duration", time.Since(startTime)))

	return nil
}
