package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

func newCredential() (azcore.TokenCredential, error) {
	return azidentity.NewDefaultAzureCredential(nil)
}

// getSubscriptions lists every enabled subscription the credential can access,
// used when the caller does not pin an explicit set with --subscriptions.
func getSubscriptions(ctx context.Context, cred azcore.TokenCredential) ([]string, error) {
	client, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create azure subscriptions client: %w", err)
	}

	subscriptions := []string{}
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list azure subscriptions: %w", err)
		}

		for _, sub := range page.Value {
			if sub.SubscriptionID == nil {
				continue
			}

			if sub.State != nil && *sub.State != armsubscriptions.SubscriptionStateEnabled {
				continue
			}

			subscriptions = append(subscriptions, *sub.SubscriptionID)
		}
	}

	return subscriptions, nil
}
