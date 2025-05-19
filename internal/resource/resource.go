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
package resource

import (
	"context"
	"fmt"

	"github.com/mikefero/osiris/internal/client"
	"go.uber.org/zap"
)

// ResourceData represents the structure of the data returned from the API
// endpoints. It contains a slice of maps, where each map represents a
// single item of data, and a name for the endpoint.
type ResourceData struct {
	// Data is a slice of maps, where each map represents a single item of data
	// returned from the API endpoint.
	Data []map[string]interface{} `json:"data"`
	// Name is the name of the endpoint from which the data was retrieved.
	// It is used to identify the source of the data in the output.
	Name string `json:"name"`
}

// Resource represents a Kong API resource with standard operations.
type Resource interface {
	// Name returns the display name of the resource
	Name() string
	// Path returns the API endpoint path for the resource
	Path() string
	// Dependencies returns a list of dependencies for the resource
	Dependencies() []string
	// List retrieves all items of the resource type
	List(ctx context.Context, client *client.Client, logger *zap.Logger) (ResourceData, error)
	// Delete removes a specific item by ID from the resource.
	Delete(ctx context.Context, client *client.Client, item map[string]interface{}, logger *zap.Logger) error
}

// BaseResource provides a basic implementation of the Resource interface
// that can be embedded in specific resource types.
type BaseResource struct {
	name         string
	path         string
	dependencies []string
}

// Name returns the display name of the resource.
func (r *BaseResource) Name() string {
	return r.name
}

// Path returns the API endpoint path for the resource.
func (r *BaseResource) Path() string {
	return r.path
}

func (r *BaseResource) Dependencies() []string {
	// Return a copy of the dependencies slice to prevent external modification
	deps := make([]string, len(r.dependencies))
	copy(deps, r.dependencies)
	return deps
}

// List retrieves all items of the resource type.
func (r *BaseResource) List(ctx context.Context, client *client.Client, logger *zap.Logger) (ResourceData, error) {
	data, err := client.GetEndpoint(ctx, r.path)
	if err != nil {
		logger.Error("error listing resource",
			zap.String("resource", r.name),
			zap.Error(err))
		return ResourceData{}, fmt.Errorf("error listing resource %s: %w", r.name, err)
	}

	if len(data) == 0 {
		logger.Debug("No data found for resource",
			zap.String("resource", r.name))
		return ResourceData{}, nil
	}

	logger.Info("Listed data for resource",
		zap.String("resource", r.name),
		zap.Int("items", len(data)))

	return ResourceData{
		Data: data,
		Name: r.name,
	}, nil
}

func (r *BaseResource) Delete(ctx context.Context, client *client.Client, item map[string]interface{},
	logger *zap.Logger,
) error {
	// Determine the ID of the item to delete
	id, ok := item["id"].(string)
	if !ok {
		name, ok := item["name"].(string)
		if !ok {
			return fmt.Errorf("invalid item format: missing id or name field")
		}
		id = name
	}

	endpointWithID := fmt.Sprintf("%s/%s", r.path, id)
	if err := client.DeleteEndpoint(ctx, endpointWithID); err != nil {
		logger.Error("error deleting resource",
			zap.String("resource", r.name),
			zap.String("id", id),
			zap.Error(err))
		return fmt.Errorf("error deleting resource %s with ID %s: %w", r.name, id, err)
	}

	logger.Debug("Deleted resource",
		zap.String("resource", r.name),
		zap.String("id", id))

	return nil
}
