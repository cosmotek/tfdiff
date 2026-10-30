package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
)

// query projects every resource down to the fields tfdiff needs. Resource Graph
// spans all requested subscriptions in a single query, so unlike the AWS path
// there is no region/resource-type fan-out.
const query = "Resources | project id, name, type, location, subscriptionId, resourceGroup"

func getResources(ctx context.Context, svc *armresourcegraph.Client, subscriptions []string, logger zerolog.Logger) ([]map[string]any, error) {
	subscriptionPtrs := lo.Map(subscriptions, func(id string, _ int) *string {
		return to.Ptr(id)
	})

	resources := []map[string]any{}
	var skipToken *string

	for {
		output, err := svc.Resources(ctx, armresourcegraph.QueryRequest{
			Query:         to.Ptr(query),
			Subscriptions: subscriptionPtrs,
			Options: &armresourcegraph.QueryRequestOptions{
				ResultFormat: to.Ptr(armresourcegraph.ResultFormatObjectArray),
				Top:          to.Ptr(int32(1000)), // this limit is set by Azure
				SkipToken:    skipToken,
			},
		}, nil)
		if err != nil {
			return nil, fmt.Errorf("azure resource graph query failed: %w", err)
		}

		rows, ok := output.Data.([]any)
		if !ok {
			return nil, fmt.Errorf("unexpected resource graph response format: %T", output.Data)
		}

		for _, row := range rows {
			if resource, ok := row.(map[string]any); ok {
				resources = append(resources, resource)
			}
		}

		logger.Debug().Int("count", len(rows)).Msg("received resource graph page")

		if output.SkipToken == nil || *output.SkipToken == "" {
			break
		}

		skipToken = output.SkipToken
	}

	return resources, nil
}
