/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package walletrunner

import (
	"encoding/json"
	"fmt"
	"strings"

	didconfigclient "github.com/trustbloc/vc-go/didconfig/client"
)

const (
	linkedDomainsService = "LinkedDomains"
)

type serviceEndpoint struct {
	Origins []string `json:"origins"`
}

func (s *Service) runLinkedDomainVerification(didID string) error {
	didDocResolution, vdrErr := s.ariesServices.vdrRegistry.Resolve(didID)
	if vdrErr != nil {
		return fmt.Errorf("failed to resolve DID %s, err: %w", didID, vdrErr)
	}

	for _, service := range didDocResolution.DIDDocument.Service {
		serviceType := getServiceType(service.Type)
		if serviceType != linkedDomainsService {
			continue
		}

		serviceEndpointBytes, err := service.ServiceEndpoint.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to get LinkedDomains service endpoint: %w", err)
		}

		serviceEndpoint := &serviceEndpoint{}

		if err := json.Unmarshal(serviceEndpointBytes, serviceEndpoint); err != nil {
			return err
		}

		didConfigurationClient := didconfigclient.New(
			didconfigclient.WithJSONLDDocumentLoader(s.ariesServices.documentLoader),
			didconfigclient.WithVDRegistry(s.ariesServices.vdrRegistry),
			didconfigclient.WithHTTPClient(s.httpClient),
		)

		if err = didConfigurationClient.VerifyDIDAndDomain(didID,
			strings.TrimSuffix(serviceEndpoint.Origins[0], "/")); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("no LinkedDomains service in DID %s", didID)
}

func getServiceType(serviceType interface{}) string {
	var val string

	switch t := serviceType.(type) {
	case string:
		val = t
	case []string:
		if len(t) > 0 {
			val = t[0]
		}
	case []interface{}:
		if len(t) > 0 {
			if str, ok := t[0].(string); ok {
				val = str
			}
		}
	}

	return val
}
