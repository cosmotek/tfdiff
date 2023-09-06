package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
)

func getResourceTypes(ctx context.Context, rsrc *resourceexplorer2.Client) ([]types.SupportedResourceType, error) {
	types := []types.SupportedResourceType{}

	var nextToken *string
	for {
		output, err := rsrc.ListSupportedResourceTypes(ctx, &resourceexplorer2.ListSupportedResourceTypesInput{
			MaxResults: aws.Int32(1000),
			NextToken:  nextToken,
		})
		if err != nil {
			return nil, err
		}

		types = append(types, output.ResourceTypes...)
		if output.NextToken == nil || *output.NextToken == "" {
			break
		}

		nextToken = output.NextToken
	}

	return types, nil
}
