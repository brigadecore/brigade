package resourceproviders

import (
	"net/http"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToResourceProviderListQuery() (string, error)
}

// ListOpts allows the filtering resource providers. Filtering is achieved by
// passing in struct field values that map to the resource provider attributes
// you want to see returned.
type ListOpts struct {
	// Name is the name of the resource provider to filter the list
	Name string `q:"name"`

	// UUID is the uuid of the resource provider to filter the list
	UUID string `q:"uuid"`

	// MemberOf is a string representing aggregate uuids to filter or exclude from the list
	MemberOf string `q:"member_of"`

	// Resources is a comma-separated list of string indicating an amount of resource
	// of a specified class that a provider must have the capacity and availability to serve
	Resources string `q:"resources"`

	// InTree is a string that represents a resource provider UUID.  The returned resource
	// providers will be in the same provider tree as the specified provider.
	InTree string `q:"in_tree"`

	// Required is comma-delimited list of string trait names.
	Required string `q:"required"`
}

// ToResourceProviderListQuery formats a ListOpts into a query string.
func (opts ListOpts) ToResourceProviderListQuery() (string, error) {
	q, err := gophercloud.BuildQueryString(opts)
	return q.String(), err
}

// List makes a request against the API to list resource providers.
func List(client *gophercloud.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := resourceProvidersListURL(client)

	if opts != nil {
		query, err := opts.ToResourceProviderListQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return ResourceProvidersPage{pagination.SinglePageBase(r)}
	})
}

// CreateOptsBuilder allows extensions to add additional parameters to the
// Create request.
type CreateOptsBuilder interface {
	ToResourceProviderCreateMap() (map[string]interface{}, error)
}

// CreateOpts represents options used to create a resource provider.
type CreateOpts struct {
	Name string `json:"name"`
	UUID string `json:"uuid,omitempty"`
}

// ToResourceProviderCreateMap constructs a request body from CreateOpts.
func (opts CreateOpts) ToResourceProviderCreateMap() (map[string]interface{}, error) {
	b, err := gophercloud.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Create makes a request against the API to create a resource provider
func Create(client *gophercloud.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToResourceProviderCreateMap()
	if err != nil {
		r.Err = err
		return
	}

	var result *http.Response

	result, r.Err = client.Post(resourceProvidersListURL(client), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})
	if r.Err == nil {
		r.Header = result.Header
	}

	return
}
