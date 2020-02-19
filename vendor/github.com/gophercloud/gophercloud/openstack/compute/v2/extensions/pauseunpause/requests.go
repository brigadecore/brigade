package pauseunpause

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions"
)

// Pause is the operation responsible for pausing a Compute server.
func Pause(client *gophercloud.ServiceClient, id string) (r PauseResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"pause": nil}, nil, nil)
	return
}

// Unpause is the operation responsible for unpausing a Compute server.
func Unpause(client *gophercloud.ServiceClient, id string) (r UnpauseResult) {
	_, r.Err = client.Post(extensions.ActionURL(client, id), map[string]interface{}{"unpause": nil}, nil, nil)
	return
}
