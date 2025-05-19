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

// CACertificateResource represents SSL CA Certificates in Kong Gateway.
type CACertificateResource struct {
	BaseResource
}

// NewCACertificate creates a new ca-certificate resource.
func NewCACertificate() Resource {
	return &CACertificateResource{
		BaseResource: BaseResource{
			name: "ca-certificate",
			path: "ca_certificates",
		},
	}
}

// List retrieves a list of CA certificates from the Kong Gateway and removes
// metadata from the response.
func (r *CACertificateResource) List(ctx context.Context, client *client.Client, logger *zap.Logger) (
	ResourceData, error,
) {
	caCertificateData, err := client.GetEndpoint(ctx, r.path)
	if err != nil {
		return ResourceData{}, fmt.Errorf("failed to list CA certificates: %w", err)
	}
	if len(caCertificateData) == 0 {
		logger.Debug("No data found for resource",
			zap.String("resource", r.name))
		return ResourceData{}, nil
	}

	// Remove metadata from CA certificates before returning
	return ResourceData{
		Data: cleanCertificateData(caCertificateData),
		Name: r.Name(),
	}, nil
}

func cleanCertificateData(certificates []map[string]interface{}) []map[string]interface{} {
	for i := range certificates {
		delete(certificates[i], "metadata")
	}
	return certificates
}
