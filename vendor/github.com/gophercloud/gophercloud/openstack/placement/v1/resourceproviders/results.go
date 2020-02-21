package resourceproviders

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

type ResourceProviderLinks struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

// ResourceProvider are entities which provider consumable inventory of one or more classes of resource
type ResourceProvider struct {
	// Generation is a consistent view marker that assists with the management of concurrent resource provider updates.
	Generation int `json:"generation"`

	// UUID of a resource provider.
	UUID string `json:"uuid"`

	// Links is a list of links associated with one resource provider.
	Links []ResourceProviderLinks `json:"links"`

	// Name of one resource provider.
	Name string `json:"name"`

	// The ParentProviderUUID contains the UUID of the immediate parent of the resource provider.
	// Requires microversion 1.14 or above
	ParentProviderUUID string `json:"parent_provider_uuid"`

	// The RootProviderUUID contains the read-only UUID of the top-most provider in this provider tree.
	// Requires microversion 1.14 or above
	RootProviderUUID string `json:"root_provider_uuid"`
}

// resourceProviderResult is the resposne of a base ResourceProvider result.
type resourceProviderResult struct {
	gophercloud.Result
}

// Extract interpets any resourceProviderResult-base result as a ResourceProvider.
func (r resourceProviderResult) Extract() (*ResourceProvider, error) {
	var s ResourceProvider
	err := r.ExtractInto(&s)

	return &s, err
}

// CreateResult is the result of a Create operation. Call its Extract
// method to interpret it as a ResourceProvider.
type CreateResult struct {
	resourceProviderResult
}

// ResourceProvidersPage contains a single page of all resource providers from a List call.
type ResourceProvidersPage struct {
	pagination.SinglePageBase
}

// IsEmpty determines if a ResourceProvidersPage contains any results.
func (page ResourceProvidersPage) IsEmpty() (bool, error) {
	resourceProviders, err := ExtractResourceProviders(page)
	return len(resourceProviders) == 0, err
}

// ExtractResourceProviders returns a slice of ResourceProvider from a List operation.
func ExtractResourceProviders(r pagination.Page) ([]ResourceProvider, error) {
	var s struct {
		ResourceProviders []ResourceProvider `json:"resource_providers"`
	}
	err := (r.(ResourceProvidersPage)).ExtractInto(&s)
	return s.ResourceProviders, err
}
