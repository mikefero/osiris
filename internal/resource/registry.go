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
	"errors"
	"fmt"
)

// Registry provides a structure for organizing and ordering resources
// based on their dependencies.
type Registry struct {
	resources []Resource
}

// orderType defines the sorting order type for resource operations.
type orderType int

const (
	// deleteOrder represents the order for deletion operations (leaf-to-root).
	deleteOrder orderType = iota
	// insertOrder represents the order for insertion operations (root-to-leaf).
	insertOrder
)

// ResourceRegistry provides a centralized collection of all Kong Gateway
// resources. This allows for a static definition of resources that the client
// can use.
var resourceRegistry = []Resource{
	NewACL(),
	NewBasicAuth(),
	NewCACertificate(),
	NewCertificate(),
	NewConfigStore(),
	NewConsumer(),
	NewConsumerGroup(),
	NewCustomPlugin(),
	NewDegraphQLRoute(),
	NewGraphQLRateLimitingAdvancedCost(),
	NewHMACAuth(),
	NewJWT(),
	NewKey(),
	NewKeyAuth(),
	NewKeySet(),
	NewMTLSAuth(),
	NewPartial(),
	NewPlugin(),
	NewPluginSchema(),
	NewRoute(),
	NewService(),
	NewSNI(),
	NewTarget(),
	NewUpstream(),
	NewVault(),
}

// NewRegistry creates a new resource registry with all predefined resources.
func NewRegistry() *Registry {
	return &Registry{
		resources: resourceRegistry,
	}
}

// GetResources returns all resources in the registry.
func (r *Registry) GetResources() []Resource {
	return r.resources
}

// GetResourcesForDeletion returns resources ordered for deletion operations.
func (r *Registry) GetResourcesForDeletion() ([][]Resource, error) {
	return r.getOrderedResources(deleteOrder)
}

func (r *Registry) getOrderedResources(orderType orderType) ([][]Resource, error) {
	// Build a map of resource names to resources for quick lookup
	resourceMap := make(map[string]Resource)
	for _, res := range r.resources {
		resourceMap[res.Name()] = res
	}

	// Build a dependency graph
	graph := make(map[string][]string)
	for _, res := range r.resources {
		name := res.Name()
		// Initialize empty dependencies list for all resources
		graph[name] = []string{}
	}

	// Add edges according to the correct direction for each order type
	for _, res := range r.resources {
		name := res.Name()
		deps := res.Dependencies()

		for _, dep := range deps {
			// Ensure the dependency exists in our resource map
			if _, exists := resourceMap[dep]; !exists {
				return nil, errors.New("dependency not found: " + dep)
			}

			switch orderType {
			case deleteOrder:
				// For deletion order: If resource R depends on D,
				// then R must be deleted BEFORE D
				// So add an edge from R -> D
				graph[name] = append(graph[name], dep)
			case insertOrder:
				// For insertion order: If resource R depends on D,
				// then D must be created BEFORE R
				// So add an edge from D -> R
				graph[dep] = append(graph[dep], name)
			default:
				return nil, errors.New("unknown order type specified")
			}
		}
	}

	// Perform topological sort
	orderedNames, err := topologicalSort(graph)
	if err != nil {
		return nil, fmt.Errorf("failed to topologically sort resources: %w", err)
	}

	// Group resources by level (for parallel execution)
	levels := make([][]Resource, len(orderedNames))
	for i, level := range orderedNames {
		levels[i] = make([]Resource, 0, len(level))
		for _, name := range level {
			if res, exists := resourceMap[name]; exists {
				levels[i] = append(levels[i], res)
			}
		}
	}

	return levels, nil
}

// topologicalSort performs a modified topological sort that groups nodes
// that can be processed in parallel at the same level.
func topologicalSort(graph map[string][]string) ([][]string, error) {
	// Initialize in-degree counts for each node
	inDegree := make(map[string]int)
	for node := range graph {
		inDegree[node] = 0
	}

	// Calculate in-degree for each node
	for _, deps := range graph {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	// Initialize queue with nodes that have no dependencies
	var queue []string
	for node := range graph {
		if inDegree[node] == 0 {
			queue = append(queue, node)
		}
	}

	// Process the graph level by level
	var result [][]string
	for len(queue) > 0 {
		// Current level contains all nodes in the queue
		level := make([]string, len(queue))
		copy(level, queue)
		result = append(result, level)

		// Create new queue for next level
		var nextQueue []string

		// Process all nodes at the current level
		for _, node := range queue {
			// Decrease in-degree of all dependent nodes
			for _, dep := range graph[node] {
				inDegree[dep]--
				// If in-degree becomes 0, add to the next level queue
				if inDegree[dep] == 0 {
					nextQueue = append(nextQueue, dep)
				}
			}
		}

		// Update queue for next level
		queue = nextQueue
	}

	// Check if all nodes were processed
	processedNodes := 0
	for _, level := range result {
		processedNodes += len(level)
	}

	if processedNodes < len(graph) {
		return nil, errors.New("cyclic dependency detected in resource graph")
	}

	return result, nil
}
