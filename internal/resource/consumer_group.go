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

// ConsumerGroupResource represents consumer groups in Kong Gateway.
type ConsumerGroupResource struct {
	BaseResource
}

// NewConsumerGroup creates a new consumer-group resource.
func NewConsumerGroup() Resource {
	return &ConsumerGroupResource{
		BaseResource: BaseResource{
			name: "consumer-group",
			path: "consumer_groups",
		},
	}
}
