package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/placement/v1/resourceproviders"

	th "github.com/gophercloud/gophercloud/testhelper"
	fake "github.com/gophercloud/gophercloud/testhelper/client"
)

const ResourceProvidersBody = `
{
  "resource_providers": [
    {
      "generation": 1,
      "uuid": "99c09379-6e52-4ef8-9a95-b9ce6f68452e",
      "links": [
        {
          "href": "/resource_providers/99c09379-6e52-4ef8-9a95-b9ce6f68452e",
          "rel": "self"
        }
      ],
      "name": "vgr.localdomain",
      "parent_provider_uuid": "542df8ed-9be2-49b9-b4db-6d3183ff8ec8",
      "root_provider_uuid": "542df8ed-9be2-49b9-b4db-6d3183ff8ec8"
    },
    {
      "generation": 2,
      "uuid": "d0b381e9-8761-42de-8e6c-bba99a96d5f5",
      "links": [
        {
          "href": "/resource_providers/d0b381e9-8761-42de-8e6c-bba99a96d5f5",
          "rel": "self"
        }
      ],
      "name": "pony1",
      "parent_provider_uuid": null,
      "root_provider_uuid": "d0b381e9-8761-42de-8e6c-bba99a96d5f5"
    }
  ]
}
`

const ResourceProviderCreateBody = `
{
  "generation": 1,
  "uuid": "99c09379-6e52-4ef8-9a95-b9ce6f68452e",
  "links": [
	{
	  "href": "/resource_providers/99c09379-6e52-4ef8-9a95-b9ce6f68452e",
	  "rel": "self"
	}
  ],
  "name": "vgr.localdomain",
  "parent_provider_uuid": "542df8ed-9be2-49b9-b4db-6d3183ff8ec8",
  "root_provider_uuid": "542df8ed-9be2-49b9-b4db-6d3183ff8ec8"
}
`

var ExpectedResourceProvider1 = resourceproviders.ResourceProvider{
	Generation: 1,
	UUID:       "99c09379-6e52-4ef8-9a95-b9ce6f68452e",
	Links: []resourceproviders.ResourceProviderLinks{
		{
			Href: "/resource_providers/99c09379-6e52-4ef8-9a95-b9ce6f68452e",
			Rel:  "self",
		},
	},
	Name:               "vgr.localdomain",
	ParentProviderUUID: "542df8ed-9be2-49b9-b4db-6d3183ff8ec8",
	RootProviderUUID:   "542df8ed-9be2-49b9-b4db-6d3183ff8ec8",
}

var ExpectedResourceProvider2 = resourceproviders.ResourceProvider{
	Generation: 2,
	UUID:       "d0b381e9-8761-42de-8e6c-bba99a96d5f5",
	Links: []resourceproviders.ResourceProviderLinks{
		{
			Href: "/resource_providers/d0b381e9-8761-42de-8e6c-bba99a96d5f5",
			Rel:  "self",
		},
	},
	Name:               "pony1",
	ParentProviderUUID: "",
	RootProviderUUID:   "d0b381e9-8761-42de-8e6c-bba99a96d5f5",
}

var ExpectedResourceProviders = []resourceproviders.ResourceProvider{
	ExpectedResourceProvider1,
	ExpectedResourceProvider2,
}

func HandleResourceProviderList(t *testing.T) {
	th.Mux.HandleFunc("/resource_providers",
		func(w http.ResponseWriter, r *http.Request) {
			th.TestMethod(t, r, "GET")
			th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			fmt.Fprintf(w, ResourceProvidersBody)
		})
}

func HandleResourceProviderCreate(t *testing.T) {
	th.Mux.HandleFunc("/resource_providers", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, ResourceProviderCreateBody)
	})
}
