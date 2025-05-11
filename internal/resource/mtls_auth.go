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

// MTLSAuthResource represents MTLS authentication credentials in Kong Gateway.
type MTLSAuthResource struct {
	BaseResource
}

// NewMTLSAuth creates a new mtls-auth resource.
func NewMTLSAuth() Resource {
	return &MTLSAuthResource{
		BaseResource: BaseResource{
			name: "mtls-auth",
			path: "mtls-auths",
		},
	}
}
