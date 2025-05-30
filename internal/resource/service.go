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

// ServiceResource represents services in Kong Gateway.
type ServiceResource struct {
	BaseResource
}

// NewService creates a new service resource.
func NewService() Resource {
	return &ServiceResource{
		BaseResource: BaseResource{
			name: "service",
			path: "services",
			dependencies: []string{
				"ca-certificate",
				"certificate",
			},
		},
	}
}
