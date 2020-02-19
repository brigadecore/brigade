package portforwarding

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

type PortForwarding struct {
	// The ID of the floating IP port forwarding
	ID string `json:"id"`

	// The ID of the Neutron port associated to the floating IP port forwarding.
	InternalPortID string `json:"internal_port_id"`

	// The TCP/UDP/other protocol port number of the port forwarding’s floating IP address.
	ExternalPort int `json:"external_port"`

	// The IP protocol used in the floating IP port forwarding.
	Protocol string `json:"protocol"`

	// The TCP/UDP/other protocol port number of the Neutron port fixed
	// IP address associated to the floating ip port forwarding.
	InternalPort int `json:"internal_port"`

	// The fixed IPv4 address of the Neutron port associated
	// to the floating IP port forwarding.
	InternalIPAddress string `json:"internal_ip_address"`
}

type commonResult struct {
	gophercloud.Result
}

// CreateResult represents the result of a create operation. Call its Extract
// method to interpret it as a PortForwarding.
type CreateResult struct {
	commonResult
}

// GetResult represents the result of a get operation. Call its Extract
// method to interpret it as a PortForwarding.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as a PortForwarding.
type UpdateResult struct {
	commonResult
}

// DeleteResult represents the result of a delete operation. Call its
// ExtractErr method to determine if the request succeeded or failed.
type DeleteResult struct {
	gophercloud.ErrResult
}

// Extract will extract a Port Forwarding resource from a result.
func (r commonResult) Extract() (*PortForwarding, error) {
	var s PortForwarding
	err := r.ExtractInto(&s)
	return &s, err
}

func (r commonResult) ExtractInto(v interface{}) error {
	return r.Result.ExtractIntoStructPtr(v, "port_forwarding")
}

// PortForwardingPage is the page returned by a pager when traversing over a
// collection of port forwardings.
type PortForwardingPage struct {
	pagination.LinkedPageBase
}

// NextPageURL is invoked when a paginated collection of port forwardings has
// reached the end of a page and the pager seeks to traverse over a new one.
// In order to do this, it needs to construct the next page's URL.
func (r PortForwardingPage) NextPageURL() (string, error) {
	var s struct {
		Links []gophercloud.Link `json:"port_forwarding_links"`
	}
	err := r.ExtractInto(&s)
	if err != nil {
		return "", err
	}
	return gophercloud.ExtractNextURL(s.Links)
}

// IsEmpty checks whether a PortForwardingPage struct is empty.
func (r PortForwardingPage) IsEmpty() (bool, error) {
	is, err := ExtractPortForwardings(r)
	return len(is) == 0, err
}

// ExtractPortForwardings accepts a Page struct, specifically a PortForwardingPage
// struct, and extracts the elements into a slice of PortForwarding structs. In
// other words, a generic collection is mapped into a relevant slice.
func ExtractPortForwardings(r pagination.Page) ([]PortForwarding, error) {
	var s struct {
		PortForwardings []PortForwarding `json:"port_forwardings"`
	}
	err := (r.(PortForwardingPage)).ExtractInto(&s)
	return s.PortForwardings, err
}
