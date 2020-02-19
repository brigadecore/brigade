/*
Package quotas provides the ability to retrieve and manage Networking quotas through the Neutron API.

Example to Get project quotas

    projectID = "23d5d3f79dfa4f73b72b8b0b0063ec55"
    quotasInfo, err := quotas.Get(networkClient, projectID).Extract()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("quotas: %#v\n", quotasInfo)

Example to Update project quotas

    projectID = "23d5d3f79dfa4f73b72b8b0b0063ec55"

    updateOpts := quotas.UpdateOpts{
        FloatingIP:        gophercloud.IntToPointer(0),
        Network:           gophercloud.IntToPointer(-1),
        Port:              gophercloud.IntToPointer(5),
        RBACPolicy:        gophercloud.IntToPointer(10),
        Router:            gophercloud.IntToPointer(15),
        SecurityGroup:     gophercloud.IntToPointer(20),
        SecurityGroupRule: gophercloud.IntToPointer(-1),
        Subnet:            gophercloud.IntToPointer(25),
        SubnetPool:        gophercloud.IntToPointer(0),
    }
    quotasInfo, err := quotas.Update(networkClient, projectID)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("quotas: %#v\n", quotasInfo)
*/
package quotas
