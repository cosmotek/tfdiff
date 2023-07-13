package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
)

func getResources(rsrc *resourceexplorer2.ResourceExplorer2, region, resourceType string) ([]*resourceexplorer2.Resource, error) {
	query := fmt.Sprintf("arn region:%s resourcetype:%s", region, resourceType)

	resources := []*resourceexplorer2.Resource{}
	var token *string

	for {
		output, err := rsrc.Search(&resourceexplorer2.SearchInput{
			MaxResults:  aws.Int64(1000), // this limit is set by AWS
			NextToken:   token,
			QueryString: aws.String(query),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case resourceexplorer2.ErrCodeUnauthorizedException:
					return nil, fmt.Errorf("Failed to query aws resource_explorer2: %w\n\nYou may need to activate resource_explorer2 in the %s region.\n", err, region)
				default:
					return nil, err
				}
			}
		}

		resources = append(resources, output.Resources...)
		if output.NextToken == nil || *output.NextToken == "" {
			break
		}

		token = output.NextToken
	}

	return resources, nil
}
