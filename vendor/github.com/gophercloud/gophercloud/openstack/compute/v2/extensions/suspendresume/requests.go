package suspendresume

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
)

// Suspend is the operation responsible for suspending a Compute server.
func Suspend(client *gophercloud.ServiceClient, id string) (r SuspendResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"suspend": nil}, nil, nil)
	return
}

// Resume is the operation responsible for resuming a Compute server.
func Resume(client *gophercloud.ServiceClient, id string) (r UnsuspendResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"resume": nil}, nil, nil)
	return
}
