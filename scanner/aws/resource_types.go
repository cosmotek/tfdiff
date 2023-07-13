package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourceexplorer2"
)

func getResourceTypes(rsrc *resourceexplorer2.ResourceExplorer2) ([]*resourceexplorer2.SupportedResourceType, error) {
	types := []*resourceexplorer2.SupportedResourceType{}

	var nextToken *string
	for {
		output, err := rsrc.ListSupportedResourceTypes(&resourceexplorer2.ListSupportedResourceTypesInput{
			MaxResults: aws.Int64(1000),
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
