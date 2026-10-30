package azure

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/cosmotek/tfdiff/scanner"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type Scanner struct {
	resourceGraphService *armresourcegraph.Client
	subscriptions        []string
	excludeResourceTypes []string

	conf   Config
	logger zerolog.Logger
}

type Config struct {
	Subscriptions        []string
	ExcludeResourceTypes []string
	MaxConcurrency       uint64
}

func New(ctx context.Context, logger zerolog.Logger, conf Config) (*Scanner, error) {
	cred, err := newCredential()
	if err != nil {
		return nil, fmt.Errorf("failed to load default azure credentials, you may need to run 'az login': %w", err)
	}

	subscriptions := conf.Subscriptions
	if len(subscriptions) == 0 {
		subscriptions, err = getSubscriptions(ctx, cred)
		if err != nil {
			return nil, err
		}

		if len(subscriptions) == 0 {
			return nil, errors.New("no accessible azure subscriptions found, you may need to log in")
		}
	}

	logger.Info().
		Strs("subscriptions", subscriptions).
		Msg("azure client instantiated")

	svc, err := armresourcegraph.NewClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure resource graph client: %w", err)
	}

	logger.Debug().Strs("excluded_types", conf.ExcludeResourceTypes).Msg("resource exclusion list specified")

	return &Scanner{
		resourceGraphService: svc,
		subscriptions:        subscriptions,
		excludeResourceTypes: conf.ExcludeResourceTypes,
		conf:                 conf,
		logger:               logger,
	}, nil
}

func (s *Scanner) RunScan(ctx context.Context) (scanner.AssetList, error) {
	rows, err := getResources(ctx, s.resourceGraphService, s.subscriptions, s.logger)
	if err != nil {
		return nil, err
	}

	assets := scanner.AssetList{}
	for _, row := range rows {
		resourceType := strings.ToLower(getString(row, "type"))

		if lo.Contains(s.excludeResourceTypes, resourceType) {
			s.logger.Debug().Str("resource_type", resourceType).Msg("dropping excluded resource type from scan")
			continue
		}

		service, shortType := splitResourceType(resourceType)
		assets = append(assets, scanner.Asset{
			Identifier:   getString(row, "id"),
			AccountID:    getString(row, "subscriptionId"),
			Region:       getString(row, "location"),
			Service:      service,
			ResourceType: shortType,
			Metadata: map[string]any{
				"name":           getString(row, "name"),
				"resource_group": getString(row, "resourceGroup"),
			},
		})
	}

	s.logger.Debug().Int("count", len(assets)).Msg("adding resources")

	return assets, nil
}

// splitResourceType divides an ARM resource type such as
// "microsoft.compute/virtualmachines" into a provider namespace and the
// remainder, mirroring the service/resource_type split used for AWS.
func splitResourceType(resourceType string) (service, shortType string) {
	parts := strings.SplitN(resourceType, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return resourceType, resourceType
}

func getString(row map[string]any, key string) string {
	if value, ok := row[key].(string); ok {
		return value
	}

	return ""
}
