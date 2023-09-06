package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cosmotek/tfdiff/scanner"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type Scanner struct {
	resourceExplorerService *resourceexplorer2.Client
	scanRegions             []string
	scanResourceTypes       []string

	conf   Config
	logger zerolog.Logger
}

type Config struct {
	ScanRegions          []string
	ExcludeResourceTypes []string
	MaxConcurrency       uint64
}

func New(ctx context.Context, logger zerolog.Logger, conf Config) (*Scanner, error) {
	if len(conf.ScanRegions) == 0 {
		conf.ScanRegions = Regions
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load default awscli config: %w", err)
	}

	client := sts.NewFromConfig(cfg)
	identity, err := client.GetCallerIdentity(
		ctx,
		&sts.GetCallerIdentityInput{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to determine awscli caller identity, you may need to log in: %w", err)
	}

	logger.Info().
		Any("aws_account", identity.Account).
		Any("user_id", identity.UserId).
		Any("arn", identity.Arn).
		Msg("aws client instantiated")

	svc := resourceexplorer2.NewFromConfig(cfg)
	rtypes, err := getResourceTypes(ctx, svc)
	if err != nil {
		return nil, err
	}

	resourceTypes := lo.Map(rtypes, func(resource types.SupportedResourceType, _ int) string {
		return *resource.ResourceType
	})

	logger.Debug().Strs("resource_types", resourceTypes).Msg("retrieved supported resource types from aws")

	logger.Debug().Strs("excluded_types", conf.ExcludeResourceTypes).Msg("resource exclusion list specified")

	resourceTypes = lo.Filter(resourceTypes, func(typeOf string, _ int) bool {
		contains := lo.Contains(conf.ExcludeResourceTypes, typeOf)
		if contains {
			logger.Debug().Str("resource_type", typeOf).Msg("dropping excluded resource type from scan")
		}

		return !contains
	})

	return &Scanner{
		resourceExplorerService: svc,
		scanRegions:             conf.ScanRegions,
		scanResourceTypes:       resourceTypes,
		conf:                    conf,
		logger:                  logger,
	}, nil
}

func (s *Scanner) RunScan(ctx context.Context) (scanner.AssetList, error) {
	assets := scanner.AssetList{}

	for _, region := range s.scanRegions {
		for _, resourceType := range s.scanResourceTypes {
			s.logger.Debug().
				Str("query", fmt.Sprintf("arn region:%s resourcetype:%s", region, resourceType)).
				Msg("querying resource explorer")

			resourcesReturned, err := getResources(ctx, s.resourceExplorerService, region, resourceType)
			if err != nil {
				return nil, err
			}

			assetsReturned := lo.Map(resourcesReturned, func(resource types.Resource, _ int) scanner.Asset {
				return scanner.Asset{
					Identifier:   *resource.Arn,
					AccountID:    *resource.OwningAccountId,
					Region:       *resource.Region,
					Service:      *resource.Service,
					ResourceType: *resource.ResourceType,
					Metadata: map[string]any{
						"props": resource.Properties,
					},
				}
			})

			s.logger.Debug().
				Str("region", region).
				Str("type", resourceType).
				Int("count", len(assetsReturned)).
				Msg("adding resources")
			assets = append(assets, assetsReturned...)

			if len(assetsReturned) == 1000 {
				s.logger.Warn().
					Str("region", region).
					Str("type", resourceType).
					Msg("query results may have been truncated by resource explorer page limits")
			}
		}
	}

	return assets, nil
}
