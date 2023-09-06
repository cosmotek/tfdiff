package aws

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2/types"
	"github.com/aws/smithy-go"
)

func getResources(ctx context.Context, rsrc *resourceexplorer2.Client, region, resourceType string) ([]types.Resource, error) {
	query := fmt.Sprintf("arn region:%s resourcetype:%s", region, resourceType)

	resources := []types.Resource{}
	var token *string

	for {
		output, err := rsrc.Search(ctx, &resourceexplorer2.SearchInput{
			MaxResults:  aws.Int32(1000), // this limit is set by AWS
			NextToken:   token,
			QueryString: aws.String(query),
		})
		if err != nil {
			var oe *smithy.OperationError
			if errors.As(err, &oe) {
				// TODO update to: "Failed to query aws resource_explorer2: %w\n\nYou may need to activate resource_explorer2 in the %s region.\n"
				return nil, fmt.Errorf("ResourceExplorer2.Search returned an error: %w", err)
			}

			return nil, fmt.Errorf("unknown error was returned: %w", err)
		}

		resources = append(resources, output.Resources...)
		if output.NextToken == nil || *output.NextToken == "" {
			break
		}

		token = output.NextToken
	}

	return resources, nil
}
