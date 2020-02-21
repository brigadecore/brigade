package quotas

import "github.com/gophercloud/gophercloud"

// Get returns Networking Quotas for a project.
func Get(client *gophercloud.ServiceClient, projectID string) (r GetResult) {
	_, r.Err = client.Get(getURL(client, projectID), &r.Body, nil)
	return
}

// UpdateOptsBuilder allows extensions to add additional parameters to the
// Update request.
type UpdateOptsBuilder interface {
	ToQuotaUpdateMap() (map[string]interface{}, error)
}

// UpdateOpts represents options used to update the Networking Quotas.
type UpdateOpts struct {
	// FloatingIP represents a number of floating IPs. A "-1" value means no limit.
	FloatingIP *int `json:"floatingip,omitempty"`

	// Network represents a number of networks. A "-1" value means no limit.
	Network *int `json:"network,omitempty"`

	// Port represents a number of ports. A "-1" value means no limit.
	Port *int `json:"port,omitempty"`

	// RBACPolicy represents a number of RBAC policies. A "-1" value means no limit.
	RBACPolicy *int `json:"rbac_policy,omitempty"`

	// Router represents a number of routers. A "-1" value means no limit.
	Router *int `json:"router,omitempty"`

	// SecurityGroup represents a number of security groups. A "-1" value means no limit.
	SecurityGroup *int `json:"security_group,omitempty"`

	// SecurityGroupRule represents a number of security group rules. A "-1" value means no limit.
	SecurityGroupRule *int `json:"security_group_rule,omitempty"`

	// Subnet represents a number of subnets. A "-1" value means no limit.
	Subnet *int `json:"subnet,omitempty"`

	// SubnetPool represents a number of subnet pools. A "-1" value means no limit.
	SubnetPool *int `json:"subnetpool,omitempty"`
}

// ToQuotaUpdateMap builds a request body from UpdateOpts.
func (opts UpdateOpts) ToQuotaUpdateMap() (map[string]interface{}, error) {
	return gophercloud.BuildRequestBody(opts, "quota")
}

// Update accepts a UpdateOpts struct and updates an existing Networking Quotas using the
// values provided.
func Update(c *gophercloud.ServiceClient, projectID string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToQuotaUpdateMap()
	if err != nil {
		r.Err = err
		return
	}
	_, r.Err = c.Put(updateURL(c, projectID), b, &r.Body, &gophercloud.RequestOpts{
		OkCodes: []int{200},
	})

	return
}
