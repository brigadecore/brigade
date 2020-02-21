package shelveunshelve

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
)

// Shelve is the operation responsible for shelving a Compute server.
func Shelve(client *gophercloud.ServiceClient, id string) (r ShelveResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"shelve": nil}, nil, nil)
	return
}

// ShelveOffload is the operation responsible for Shelve-Offload a Compute server.
func ShelveOffload(client *gophercloud.ServiceClient, id string) (r ShelveOffloadResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"shelveOffload": nil}, nil, nil)
	return
}

// UnshelveOptsBuilder allows extensions to add additional parameters to the
// Unshelve request.
type UnshelveOptsBuilder interface {
	ToUnshelveMap() (map[string]interface{}, error)
}

// UnshelveOpts specifies parameters of shelve-offload action.
type UnshelveOpts struct {
	// Sets the availability zone to unshelve a server
	// Available only after nova 2.77
	AvailabilityZone string `json:"availability_zone,omitempty"`
}

func (opts UnshelveOpts) ToUnshelveMap() (map[string]interface{}, error) {
	// Key 'availabilty_zone' is required if the unshelve action is an object
	// i.e {"unshelve": {}} will be rejected
	b, err := gophercloud.BuildRequestBody(opts, "unshelve")
	if err != nil {
		return nil, err
	}

	if _, ok := b["unshelve"].(map[string]interface{})["availability_zone"]; !ok {
		b["unshelve"] = nil
	}

	return b, err
}

// Unshelve is the operation responsible for unshelve a Compute server.
func Unshelve(client *gophercloud.ServiceClient, id string, opts UnshelveOptsBuilder) (r UnshelveResult) {
	b, err := opts.ToUnshelveMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = client.Post(extensions.ActionURL(client, id), b, nil, nil)
	return
}
