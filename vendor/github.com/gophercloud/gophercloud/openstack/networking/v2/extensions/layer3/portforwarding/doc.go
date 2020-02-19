/*
package portforwarding enables management and retrieval of port forwarding resources for Floating IPs from the
OpenStack Networking service.

Example to list all Port Forwardings for a floating IP

	fipID := "2f245a7b-796b-4f26-9cf9-9e82d248fda7"
	allPages, err := portforwarding.List(client, portforwarding.ListOpts{}, fipID).AllPages()
	if err != nil {
		panic(err)
	}

	allPFs, err := portforwarding.ExtractPortForwardings(allPages)
	if err != nil {
		panic(err)
	}

	for _, pf := range allPFs {
		fmt.Printf("%+v\n", pf)
	}

Example to Get a Port Forwarding with a certain ID

	fipID := "2f245a7b-796b-4f26-9cf9-9e82d248fda7"
	pfID := "725ade3c-9760-4880-8080-8fc2dbab9acc"
	pf, err := portforwarding.Get(client, fipID, pfID).Extract()
	if err != nil {
		panic(err)
	}


Example to Create a Port Forwarding for a floating IP

	createOpts := &portforwarding.CreateOpts{
		Protocol:          "tcp",
		InternalPort:      25,
		ExternalPort:      2230,
		InternalIPAddress: internalIP,
		InternalPortID:    portID,
	}

	pf, err := portforwarding.Create(networkingClient, floatingIPID, createOpts).Extract()

	if err != nil {
		panic(err)
	}

Example to Update a Port Forwarding

	updateOpts := portforwarding.UpdateOpts{
		Protocol:     "udp",
		InternalPort: 30,
		ExternalPort: 678,
	}
	fipID := "2f245a7b-796b-4f26-9cf9-9e82d248fda7"
	pfID := "725ade3c-9760-4880-8080-8fc2dbab9acc"

	pf, err := portforwarding.Update(client, fipID, pfID, updateOpts).Extract()
	if err != nil {
		panic(err)
	}

Example to Delete a Port forwarding

	fipID := "2f245a7b-796b-4f26-9cf9-9e82d248fda7"
	pfID := "725ade3c-9760-4880-8080-8fc2dbab9acc"
	err := portforwarding.Delete(networkClient, fipID, pfID).ExtractErr()
	if err != nil {
		panic(err)
	}
*/
package portforwarding
