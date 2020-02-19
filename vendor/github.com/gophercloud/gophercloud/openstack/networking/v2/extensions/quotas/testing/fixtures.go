package testing

import "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/quotas"

const GetResponseRaw = `
{
    "quota": {
        "floatingip": 15,
        "network": 20,
        "port": 25,
        "rbac_policy": -1,
        "router": 30,
        "security_group": 35,
        "security_group_rule": 40,
        "subnet": 45,
        "subnetpool": -1
    }
}
`

var GetResponse = quotas.Quota{
	FloatingIP:        15,
	Network:           20,
	Port:              25,
	RBACPolicy:        -1,
	Router:            30,
	SecurityGroup:     35,
	SecurityGroupRule: 40,
	Subnet:            45,
	SubnetPool:        -1,
}

const UpdateRequestResponseRaw = `
{
    "quota": {
        "floatingip": 0,
        "network": -1,
        "port": 5,
        "rbac_policy": 10,
        "router": 15,
        "security_group": 20,
        "security_group_rule": -1,
        "subnet": 25,
        "subnetpool": 0
    }
}
`

var UpdateResponse = quotas.Quota{
	FloatingIP:        0,
	Network:           -1,
	Port:              5,
	RBACPolicy:        10,
	Router:            15,
	SecurityGroup:     20,
	SecurityGroupRule: -1,
	Subnet:            25,
	SubnetPool:        0,
}
