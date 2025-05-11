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

import "context"

// ResourceData represents the structure of the data returned from the API
// endpoints. It contains a slice of maps, where each map represents a
// single item of data, and a name for the endpoint.
type ResourceData struct {
	// Data is a slice of maps, where each map represents a single item of data
	// returned from the API endpoint.
	Data []map[string]interface{} `json:"data"`
	// Name is the name of the endpoint from which the data was fetched.
	// It is used to identify the source of the data in the output.
	Name string `json:"name"`
}

// APIClient defines the methods a resource needs from the API client
// to perform its operations. This interface allows resources to operate
// without depending on the full client implementation.
type APIClient interface {
	// FetchData fetches data from the specified endpoint and returns
	// it as a slice of maps.
	FetchData(ctx context.Context, endpoint string) ([]map[string]interface{}, error)
}

// Resource represents a Kong API resource with standard operations.
type Resource interface {
	// Name returns the display name of the resource
	Name() string
	// Path returns the API endpoint path for the resource
	Path() string
	// List fetches all items of the resource type
	List(ctx context.Context, client APIClient) ([]map[string]interface{}, error)
}

// BaseResource provides a basic implementation of the Resource interface
// that can be embedded in specific resource types.
type BaseResource struct {
	name string
	path string
}

// Name returns the display name of the resource.
func (r *BaseResource) Name() string {
	return r.name
}

// Path returns the API endpoint path for the resource.
func (r *BaseResource) Path() string {
	return r.path
}

// List fetches all items of the resource type.
func (r *BaseResource) List(ctx context.Context, client APIClient) ([]map[string]interface{}, error) {
	return client.FetchData(ctx, r.path)
}

// New creates a new BaseResource with the given name and path.
func New(name, path string) Resource {
	return &BaseResource{
		name: name,
		path: path,
	}
}
