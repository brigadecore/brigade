package layer3

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	networking "github.com/gophercloud/gophercloud/acceptance/openstack/networking/v2"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/portforwarding"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestLayer3PortForwardingsCreateDelete(t *testing.T) {
	clients.RequirePortForwarding(t)

	client, err := clients.NewNetworkV2Client()
	th.AssertNoErr(t, err)

	choices, err := clients.AcceptanceTestChoicesFromEnv()
	th.AssertNoErr(t, err)

	// Create Network
	network, err := networking.CreateNetwork(t, client)
	th.AssertNoErr(t, err)
	defer networking.DeleteNetwork(t, client, network.ID)

	subnet, err := networking.CreateSubnet(t, client, network.ID)
	th.AssertNoErr(t, err)
	defer networking.DeleteSubnet(t, client, subnet.ID)

	router, err := CreateExternalRouter(t, client)
	th.AssertNoErr(t, err)
	defer DeleteRouter(t, client, router.ID)

	port, err := networking.CreatePort(t, client, network.ID, subnet.ID)
	th.AssertNoErr(t, err)
	// not required, since "DeleteRouterInterface" below removes the port
	// defer networking.DeletePort(t, client, port.ID)

	_, err = CreateRouterInterface(t, client, port.ID, router.ID)
	th.AssertNoErr(t, err)
	defer DeleteRouterInterface(t, client, port.ID, router.ID)

	fip, err := CreateFloatingIP(t, client, choices.ExternalNetworkID, "")
	th.AssertNoErr(t, err)
	defer DeleteFloatingIP(t, client, fip.ID)

	newFip, err := floatingips.Get(client, fip.ID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, newFip)

	pf, err := CreatePortForwarding(t, client, fip.ID, port.ID, port.FixedIPs)
	th.AssertNoErr(t, err)
	defer DeletePortForwarding(t, client, fip.ID, pf.ID)
	tools.PrintResource(t, pf)

	newPf, err := portforwarding.Get(client, fip.ID, pf.ID).Extract()
	th.AssertNoErr(t, err)

	updateOpts := portforwarding.UpdateOpts{
		Protocol:     "udp",
		InternalPort: 30,
		ExternalPort: 678,
	}

	_, err = portforwarding.Update(client, fip.ID, newPf.ID, updateOpts).Extract()
	th.AssertNoErr(t, err)

	newPf, err = portforwarding.Get(client, fip.ID, pf.ID).Extract()
	th.AssertNoErr(t, err)

	allPages, err := portforwarding.List(client, portforwarding.ListOpts{}, fip.ID).AllPages()
	th.AssertNoErr(t, err)

	allPFs, err := portforwarding.ExtractPortForwardings(allPages)
	th.AssertNoErr(t, err)

	var found bool
	for _, pf := range allPFs {
		if pf.ID == newPf.ID {
			found = true
		}
	}

	th.AssertEquals(t, true, found)

}
