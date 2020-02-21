package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/providers"
	th "github.com/gophercloud/gophercloud/testhelper"
	"github.com/gophercloud/gophercloud/testhelper/client"
)

// ProvidersListBody contains the canned body of a provider list response.
const ProvidersListBody = `
{
	"providers":[
	         {
			"name": "amphora",
			"description": "The Octavia Amphora driver."
		},
		{
			"name": "ovn",
			"description": "The Octavia OVN driver"
		}
	]
}
`

var (
	ProviderAmphora = providers.Provider{
		Name:        "amphora",
		Description: "The Octavia Amphora driver.",
	}
	ProviderOVN = providers.Provider{
		Name:        "ovn",
		Description: "The Octavia OVN driver",
	}
)

// HandleProviderListSuccessfully sets up the test server to respond to a provider List request.
func HandleProviderListSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/v2.0/lbaas/providers", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")
		r.ParseForm()
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			fmt.Fprintf(w, ProvidersListBody)
		default:
			t.Fatalf("/v2.0/lbaas/providers invoked with unexpected marker=[%s]", marker)
		}
	})
}
