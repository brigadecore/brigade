package quotas

import "github.com/gophercloud/gophercloud"

type commonResult struct {
	gophercloud.Result
}

// Extract is a function that accepts a result and extracts a Quota resource.
func (r commonResult) Extract() (*Quota, error) {
	var s struct {
		Quota *Quota `json:"quota"`
	}
	err := r.ExtractInto(&s)
	return s.Quota, err
}

// GetResult represents the result of a get operation. Call its Extract
// method to interpret it as a Quota.
type GetResult struct {
	commonResult
}

// UpdateResult represents the result of an update operation. Call its Extract
// method to interpret it as a Quota.
type UpdateResult struct {
	commonResult
}

// Quota contains Networking quotas for a project.
type Quota struct {
	// FloatingIP represents a number of floating IPs. A "-1" value means no limit.
	FloatingIP int `json:"floatingip"`

	// Network represents a number of networks. A "-1" value means no limit.
	Network int `json:"network"`

	// Port represents a number of ports. A "-1" value means no limit.
	Port int `json:"port"`

	// RBACPolicy represents a number of RBAC policies. A "-1" value means no limit.
	RBACPolicy int `json:"rbac_policy"`

	// Router represents a number of routers. A "-1" value means no limit.
	Router int `json:"router"`

	// SecurityGroup represents a number of security groups. A "-1" value means no limit.
	SecurityGroup int `json:"security_group"`

	// SecurityGroupRule represents a number of security group rules. A "-1" value means no limit.
	SecurityGroupRule int `json:"security_group_rule"`

	// Subnet represents a number of subnets. A "-1" value means no limit.
	Subnet int `json:"subnet"`

	// SubnetPool represents a number of subnet pools. A "-1" value means no limit.
	SubnetPool int `json:"subnetpool"`
}
