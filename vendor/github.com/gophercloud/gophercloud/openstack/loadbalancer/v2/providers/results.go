package providers

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// Provider is the Octavia driver that implements the load balancing mechanism
type Provider struct {
	// Human-readable description for the Loadbalancer.
	Description string `json:"description"`

	// Human-readable name for the Provider.
	Name string `json:"name"`
}

// ProviderPage is the page returned by a pager when traversing over a
// collection of providers.
type ProviderPage struct {
	pagination.LinkedPageBase
}

// NextPageURL is invoked when a paginated collection of providers has
// reached the end of a page and the pager seeks to traverse over a new one.
// In order to do this, it needs to construct the next page's URL.
func (r ProviderPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"providers_links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// IsEmpty checks whether a ProviderPage struct is empty.
func (r ProviderPage) IsEmpty() (bool, error) {
	is, err := ExtractProviders(r)
	return len(is) == 0, err
}

// ExtractProviders accepts a Page struct, specifically a ProviderPage
// struct, and extracts the elements into a slice of Provider structs. In
// other words, a generic collection is mapped into a relevant slice.
func ExtractProviders(r pagination.Page) ([]Provider, error) {
	var s struct {
		Providers []Provider `json:"providers"`
	}
	err := (r.(ProviderPage)).ExtractInto(&s)
	return s.Providers, err
}

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a provider.
func (r commonResult) Extract() (*Provider, error) {
	var s struct {
		Provider *Provider `json:"provider"`
	}
	err := r.ExtractInto(&s)
	return s.Provider, err
}

// GetResult represents the result of a get operation. Call its Extract
// method to interpret it as a Provider.
type GetResult struct {
	commonResult
}
