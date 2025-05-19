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

// CertificateResource represents SSL Certificates in Kong Gateway.
type CertificateResource struct {
	BaseResource
}

// NewCertificate creates a new certificate resource.
func NewCertificate() Resource {
	return &CertificateResource{
		BaseResource: BaseResource{
			name: "certificate",
			path: "certificates",
		},
	}
}

// List retrieves a list of certificates from the Kong Gateway and removes
// metadata from the response.
func (r *CertificateResource) List(ctx context.Context, client *client.Client, logger *zap.Logger) (
	ResourceData, error,
) {
	certificateData, err := client.GetEndpoint(ctx, r.path)
	if err != nil {
		return ResourceData{}, fmt.Errorf("failed to list certificates: %w", err)
	}
	if len(certificateData) == 0 {
		logger.Debug("No data found for resource",
			zap.String("resource", r.name))
		return ResourceData{}, nil
	}

	// Remove metadata from certificates before returning
	return ResourceData{
		Data: cleanCertificateData(certificateData),
		Name: r.Name(),
	}, nil
}
