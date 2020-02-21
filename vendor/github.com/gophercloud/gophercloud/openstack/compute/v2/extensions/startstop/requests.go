package startstop

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
)

// Start is the operation responsible for starting a Compute server.
func Start(client *gophercloud.ServiceClient, id string) (r StartResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"os-start": nil}, nil, nil)
	return
}

// Stop is the operation responsible for stopping a Compute server.
func Stop(client *gophercloud.ServiceClient, id string) (r StopResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"os-stop": nil}, nil, nil)
	return
}
