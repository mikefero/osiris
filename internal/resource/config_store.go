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
)

// ConfigStoreResource represents config stores in Konnect Only.
type ConfigStoreResource struct {
	BaseResource
}

// NewConfigStore creates a new config-store resource.
func NewConfigStore() Resource {
	return &ConfigStoreResource{
		BaseResource: BaseResource{
			name: "config-store",
			path: "config-stores",
		},
	}
}

// List retrieves a list of config stores and secrets from Konnect.
func (r *ConfigStoreResource) List(ctx context.Context, client APIClient) ([]map[string]interface{}, error) {
	configStoreData, err := client.FetchData(ctx, r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config stores: %w", err)
	}

	// Gather consumer IDs to determine if they are part of a consumer group
	for i, configStore := range configStoreData {
		id, ok := configStore["id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid config store ID for item %d", i)
		}

		// Fetch secrets keys for this config store since the values are not
		// returned in the list
		secretsPath := fmt.Sprintf("%s/%s/secrets", r.path, id)
		secrets, _ := client.FetchData(ctx, secretsPath)
		if len(secrets) > 0 {
			secretKeys := make([]string, len(secrets))
			for j, secret := range secrets {
				secretKey, ok := secret["key"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid secret key for item %d in config store %d", i, j)
				}
				secretKeys[j] = secretKey
			}
			configStore["secret"] = secretKeys
		}

		// Update the config store data with the modified config store
		configStoreData[i] = configStore
	}

	return configStoreData, nil
}
