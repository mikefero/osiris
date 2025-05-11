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

// ConsumerResource represents consumers in Kong Gateway.
type ConsumerResource struct {
	BaseResource
}

// NewConsumer creates a new consumer resource.
func NewConsumer() Resource {
	return &ConsumerResource{
		BaseResource: BaseResource{
			name: "consumer",
			path: "consumers",
		},
	}
}

// List retrieves a list of consumers from the Kong Gateway and includes their
// associated consumer groups.
func (r *ConsumerResource) List(ctx context.Context, client APIClient) ([]map[string]interface{}, error) {
	consumerData, err := client.FetchData(ctx, r.path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch consumer resource: %w", err)
	}

	// Gather consumer IDs to determine if they are part of a consumer group
	for i, consumer := range consumerData {
		id, ok := consumer["id"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid consumer ID for item %d", i)
		}

		// Fetch consumer group IDs for this consumer
		consumerGroupsPath := fmt.Sprintf("%s/%s/consumer_groups", r.path, id)
		consumerGroups, _ := client.FetchData(ctx, consumerGroupsPath)
		if len(consumerGroups) > 0 {
			consumerGroupIDs := make([]string, len(consumerGroups))
			for j, group := range consumerGroups {
				groupID, ok := group["id"].(string)
				if !ok {
					return nil, fmt.Errorf("invalid consumer group ID for item %d in consumer group %d", i, j)
				}
				consumerGroupIDs[j] = groupID
			}
			consumer["groups"] = consumerGroupIDs
		}

		// Update the consumer data with the modified consumer
		consumerData[i] = consumer
	}

	return consumerData, nil
}
