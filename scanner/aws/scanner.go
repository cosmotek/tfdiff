package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
	"github.com/cosmotek/tfdiff/scanner"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

type Scanner struct {
	resourceExplorerService *resourceexplorer2.ResourceExplorer2
	scanRegions             []string
	scanResourceTypes       []string

	conf   Config
	logger zerolog.Logger
}

type Config struct {
	ScanRegions               []string `split_words:"true" required:"false"`
	ResourceExplorerAWSRegion string   `split_words:"true" required:"false" default:"us-east-1"`
	MaxConcurrency            uint64   `split_words:"true" required:"false" default:"2"`
}

func New(logger zerolog.Logger, conf Config) (*Scanner, error) {
	if len(conf.ScanRegions) == 0 {
		conf.ScanRegions = Regions
	}

	session, err := session.NewSession(&aws.Config{
		Region: aws.String(conf.ResourceExplorerAWSRegion),
	})
	if err != nil {
		return nil, err
	}

	svc := resourceexplorer2.New(session)
	types, err := getResourceTypes(svc)
	if err != nil {
		return nil, err
	}

	resourceTypes := lo.Map(types, func(resource *resourceexplorer2.SupportedResourceType, _ int) string {
		return *resource.ResourceType
	})

	return &Scanner{
		resourceExplorerService: svc,
		scanRegions:             conf.ScanRegions,
		scanResourceTypes:       resourceTypes,
		conf:                    conf,
		logger:                  logger,
	}, nil
}

func (s *Scanner) RunScan() (scanner.AssetList, error) {
	assets := scanner.AssetList{}

	for _, region := range s.scanRegions {
		for _, resourceType := range s.scanResourceTypes {
			s.logger.Debug().
				Str("query", fmt.Sprintf("arn region:%s resourcetype:%s", region, resourceType)).
				Msg("querying resource explorer")

			resourcesReturned, err := getResources(s.resourceExplorerService, region, resourceType)
			if err != nil {
				return nil, err
			}

			assetsReturned := lo.Map(resourcesReturned, func(resource *resourceexplorer2.Resource, _ int) scanner.Asset {
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
