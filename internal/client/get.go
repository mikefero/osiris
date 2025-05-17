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
	"strings"
	"time"

	"go.uber.org/zap"
)

// GetEndpoint retrieves all data from a specified endpoint, handling
// pagination and rate limiting. It returns a slice of maps containing the
// data from the endpoint, or an error if the request fails.
func (c *Client) GetEndpoint(ctx context.Context, endpoint string) ([]map[string]interface{}, error) {
	endpointURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	var result []map[string]interface{}

	c.logger.Debug("Getting endpoint",
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
		c.logger.Debug("Getting page",
			zap.String("endpoint", endpoint),
			zap.String("page-url", pageURL),
			zap.Int("page-number", pageCount))

		data, nextPageURL, err := c.getEndpointPage(ctx, pageURL)
		if err != nil {
			// Check if the error is a RateLimitError
			errRateLimit, ok := err.(*RateLimitError)
			if !ok {
				return nil, fmt.Errorf("error getting endpoint %s: %w", endpoint, err)
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

		c.logger.Debug("Retrieved data from page",
			zap.String("endpoint", endpoint),
			zap.String("page-url", pageURL),
			zap.Int("page-number", pageCount),
			zap.Int("item-count", len(data)),
			zap.Duration("request-duration", time.Since(requestStartTime)))

		result = append(result, data...)

		if len(nextPageURL) == 0 {
			c.logger.Debug("No more pages to get",
				zap.String("endpoint", endpoint),
				zap.String("page-url", pageURL))
			break
		}
		pageURL = nextPageURL
	}

	c.logger.Debug("Retrieved all pages",
		zap.String("endpoint", endpoint),
		zap.Int("total-pages", pageCount),
		zap.Int("total-items", len(result)),
		zap.Duration("get-duration", time.Since(startTime)))

	return result, nil
}

func (c *Client) getEndpointPage(ctx context.Context, url string) ([]map[string]interface{}, string, error) {
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
		pageResp := struct {
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

		// Determine the next URL to request
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
		retryDuration := c.retryAfterDuration(resp)
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
