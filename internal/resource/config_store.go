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
func (r *ConfigStoreResource) List(ctx context.Context, client *client.Client, logger *zap.Logger) (
	ResourceData, error,
) {
	configStoreData, err := client.GetEndpoint(ctx, r.path)
	if err != nil {
		return ResourceData{}, fmt.Errorf("failed to list config stores: %w", err)
	}
	if len(configStoreData) == 0 {
		logger.Debug("No data found for resource",
			zap.String("resource", r.name))
		return ResourceData{}, nil
	}

	// Gather consumer IDs to determine if they are part of a consumer group
	for i, configStore := range configStoreData {
		id, ok := configStore["id"].(string)
		if !ok {
			return ResourceData{}, fmt.Errorf("invalid config store ID for item %d", i)
		}

		// List secrets keys for this config store since the values are not
		// returned in the list
		secretsPath := fmt.Sprintf("%s/%s/secrets", r.path, id)
		secrets, _ := client.GetEndpoint(ctx, secretsPath)
		if len(secrets) > 0 {
			secretKeys := make([]string, len(secrets))
			for j, secret := range secrets {
				secretKey, ok := secret["key"].(string)
				if !ok {
					return ResourceData{}, fmt.Errorf("invalid secret key for item %d in config store %d", i, j)
				}
				secretKeys[j] = secretKey
			}
			configStore["secret"] = secretKeys
		}

		// Update the config store data with the modified config store
		configStoreData[i] = configStore
	}

	return ResourceData{
		Data: configStoreData,
		Name: r.Name(),
	}, nil
}

func (r *ConfigStoreResource) Delete(ctx context.Context, client *client.Client, item map[string]interface{},
	logger *zap.Logger,
) error {
	id, ok := item["id"].(string)
	if !ok {
		return fmt.Errorf("invalid config store ID: %v", item)
	}

	// Delete secrets keys for this config store
	secrets, ok := item["secret"].([]string)
	if ok && len(secrets) > 0 {
		for _, secretKey := range secrets {
			// Construct the path to delete the secret
			secretPath := fmt.Sprintf("%s/%s/secrets/%s", r.path, id, secretKey)
			if err := client.DeleteEndpoint(ctx, secretPath); err != nil {
				return fmt.Errorf("failed to delete secret %s for config store %s: %w", secretKey, id, err)
			}
		}
	}

	// Delete the config store by ID
	path := fmt.Sprintf("%s/%s", r.path, id)
	if err := client.DeleteEndpoint(ctx, path); err != nil {
		return fmt.Errorf("failed to delete config store %s: %w", id, err)
	}

	return nil
}
