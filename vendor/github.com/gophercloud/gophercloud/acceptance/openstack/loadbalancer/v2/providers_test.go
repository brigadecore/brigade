// +build acceptance networking loadbalancer providers

package v2

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/providers"
)

func TestProvidersList(t *testing.T) {
	clients.SkipRelease(t, "stable/mitaka")
	clients.SkipRelease(t, "stable/newton")
	clients.SkipRelease(t, "stable/ocata")
	clients.SkipRelease(t, "stable/pike")
	clients.SkipRelease(t, "stable/queens")
	clients.SkipRelease(t, "stable/rocky")

	client, err := clients.NewLoadBalancerV2Client()
	if err != nil {
		t.Fatalf("Unable to create a loadbalancer client: %v", err)
	}

	allPages, err := providers.List(client, nil).AllPages()
	if err != nil {
		t.Fatalf("Unable to list providers: %v", err)
	}

	allProviders, err := providers.ExtractProviders(allPages)
	if err != nil {
		t.Fatalf("Unable to extract providers: %v", err)
	}

	for _, provider := range allProviders {
		tools.PrintResource(t, provider)
	}
}
