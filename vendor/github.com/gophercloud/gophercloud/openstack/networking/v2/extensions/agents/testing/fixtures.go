package testing

import (
	"time"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/agents"
)

// AgentsListResult represents raw response for the List request.
const AgentsListResult = `
{
    "agents": [
        {
            "admin_state_up": true,
            "agent_type": "Open vSwitch agent",
            "alive": true,
            "availability_zone": null,
            "binary": "neutron-openvswitch-agent",
            "configurations": {
                "datapath_type": "system",
                "extensions": [
                    "qos"
                ]
            },
            "created_at": "2017-07-26 23:15:44",
            "description": null,
            "heartbeat_timestamp": "2019-01-09 10:28:53",
            "host": "compute1",
            "id": "59186d7b-b512-4fdf-bbaf-5804ffde8811",
            "started_at": "2018-06-26 21:46:19",
            "topic": "N/A"
        },
        {
            "admin_state_up": true,
            "agent_type": "Open vSwitch agent",
            "alive": true,
            "availability_zone": null,
            "binary": "neutron-openvswitch-agent",
            "configurations": {
                "datapath_type": "system",
                "extensions": [
                    "qos"
                ]
            },
            "created_at": "2017-01-22 14:00:50",
            "description": null,
            "heartbeat_timestamp": "2019-01-09 10:28:50",
            "host": "compute2",
            "id": "76af7b1f-d61b-4526-94f7-d2e14e2698df",
            "started_at": "2018-11-06 12:09:17",
            "topic": "N/A"
        }
    ]
}
`

// Agent1 represents first unmarshalled address scope from the
// AgentsListResult.
var Agent1 = agents.Agent{
	ID:           "59186d7b-b512-4fdf-bbaf-5804ffde8811",
	AdminStateUp: true,
	AgentType:    "Open vSwitch agent",
	Alive:        true,
	Binary:       "neutron-openvswitch-agent",
	Configurations: map[string]interface{}{
		"datapath_type": "system",
		"extensions": []interface{}{
			"qos",
		},
	},
	CreatedAt:          time.Date(2017, 7, 26, 23, 15, 44, 0, time.UTC),
	StartedAt:          time.Date(2018, 6, 26, 21, 46, 19, 0, time.UTC),
	HeartbeatTimestamp: time.Date(2019, 1, 9, 10, 28, 53, 0, time.UTC),
	Host:               "compute1",
	Topic:              "N/A",
}

// Agent2 represents second unmarshalled address scope from the
// AgentsListResult.
var Agent2 = agents.Agent{
	ID:           "76af7b1f-d61b-4526-94f7-d2e14e2698df",
	AdminStateUp: true,
	AgentType:    "Open vSwitch agent",
	Alive:        true,
	Binary:       "neutron-openvswitch-agent",
	Configurations: map[string]interface{}{
		"datapath_type": "system",
		"extensions": []interface{}{
			"qos",
		},
	},
	CreatedAt:          time.Date(2017, 1, 22, 14, 00, 50, 0, time.UTC),
	StartedAt:          time.Date(2018, 11, 6, 12, 9, 17, 0, time.UTC),
	HeartbeatTimestamp: time.Date(2019, 1, 9, 10, 28, 50, 0, time.UTC),
	Host:               "compute2",
	Topic:              "N/A",
}

// AgentsGetResult represents raw response for the Get request.
const AgentsGetResult = `
{
    "agent": {
        "binary": "neutron-openvswitch-agent",
        "description": null,
        "availability_zone": null,
        "heartbeat_timestamp": "2019-01-09 11:43:01",
        "admin_state_up": true,
        "alive": true,
        "id": "43583cf5-472e-4dc8-af5b-6aed4c94ee3a",
        "topic": "N/A",
        "host": "compute3",
        "agent_type": "Open vSwitch agent",
        "started_at": "2018-06-26 21:46:20",
        "created_at": "2017-07-26 23:02:05",
        "configurations": {
            "ovs_hybrid_plug": false,
            "datapath_type": "system",
            "vhostuser_socket_dir": "/var/run/openvswitch",
            "log_agent_heartbeats": false,
            "l2_population": true,
            "enable_distributed_routing": false
        }
    }
}
`

// AgentDHCPNetworksListResult represents raw response for the ListDHCPNetworks request.
const AgentDHCPNetworksListResult = `
{
    "networks": [
        {
            "admin_state_up": true,
            "availability_zone_hints": [],
            "availability_zones": [
                "nova"
            ],
            "created_at": "2016-03-08T20:19:41",
            "dns_domain": "my-domain.org.",
            "id": "d32019d3-bc6e-4319-9c1d-6722fc136a22",
            "ipv4_address_scope": null,
            "ipv6_address_scope": null,
            "l2_adjacency": false,
            "mtu": 1500,
            "name": "net1",
            "port_security_enabled": true,
            "project_id": "4fd44f30292945e481c7b8a0c8908869",
            "qos_policy_id": "6a8454ade84346f59e8d40665f878b2e",
            "revision_number": 1,
            "router:external": false,
            "shared": false,
            "status": "ACTIVE",
            "subnets": [
                "54d6f61d-db07-451c-9ab3-b9609b6b6f0b"
            ],
            "tenant_id": "4fd44f30292945e481c7b8a0c8908869",
            "updated_at": "2016-03-08T20:19:41",
            "vlan_transparent": true,
            "description": "",
            "is_default": false
        }
    ]
}
`
