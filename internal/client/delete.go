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
	"time"

	"go.uber.org/zap"
)

// DeleteEndpoint deletes an item from the specified resource endpoint while
// handling rate limiting. It returns an error if the deletion fails or if the
// status code is not 204 No Content.
func (c *Client) DeleteEndpoint(ctx context.Context, endpointWithID string) error {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpointWithID)

	// Keep trying until successful or an error occurs
	for {
		if err := ctx.Err(); err != nil {
			c.logger.Warn("Context canceled during delete operation",
				zap.String("url", url),
				zap.Error(err))
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
		if err != nil {
			return fmt.Errorf("error creating request: %w", err)
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
			return fmt.Errorf("error making request: %w", err)
		}
		//nolint: errcheck
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusNoContent:
			c.logger.Debug("Deleted item",
				zap.String("url", url),
				zap.Duration("request-duration", time.Since(startTime)))
			return nil
		case http.StatusTooManyRequests:
			retryDuration := c.retryAfterDuration(resp)
			c.logger.Warn("Rate limit exceeded; retrying",
				zap.String("url", url),
				zap.Duration("retry-after", retryDuration))
			time.Sleep(retryDuration)
			continue
		default:
			c.logger.Error("error deleting item",
				zap.String("url", url),
				zap.Int("status-code", resp.StatusCode))
			return fmt.Errorf("unable to delete item %s: status code %d", endpointWithID, resp.StatusCode)
		}
	}
}
