// +build acceptance containerinfra

package v1

import (
	"fmt"
	"testing"
	"time"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/containerinfra/v1/nodegroups"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestNodeGroupsCRUD(t *testing.T) {
	// API not available until Magnum train
	clients.SkipRelease(t, "stable/mitaka")
	clients.SkipRelease(t, "stable/newton")
	clients.SkipRelease(t, "stable/ocata")
	clients.SkipRelease(t, "stable/pike")
	clients.SkipRelease(t, "stable/queens")
	clients.SkipRelease(t, "stable/rocky")
	clients.SkipRelease(t, "stable/stein")

	client, err := clients.NewContainerInfraV1Client()
	th.AssertNoErr(t, err)

	client.Microversion = "1.9"

	clusterTemplate, err := CreateKubernetesClusterTemplate(t, client)
	th.AssertNoErr(t, err)
	defer DeleteClusterTemplate(t, client, clusterTemplate.UUID)

	clusterID, err := CreateKubernetesCluster(t, client, clusterTemplate.UUID)
	th.AssertNoErr(t, err)
	defer DeleteCluster(t, client, clusterID)

	var nodeGroupID string

	t.Run("list", func(t *testing.T) { testNodeGroupsList(t, client, clusterID) })
	t.Run("listone-get", func(t *testing.T) { testNodeGroupGet(t, client, clusterID) })
	t.Run("create", func(t *testing.T) { nodeGroupID = testNodeGroupCreate(t, client, clusterID) })

	t.Logf("Created nodegroup: %s", nodeGroupID)

	// Wait for the node group to finish creating
	err = tools.WaitForTimeout(func() (bool, error) {
		ng, err := nodegroups.Get(client, clusterID, nodeGroupID).Extract()
		if err != nil {
			return false, fmt.Errorf("error waiting for node group to create: %v", err)
		}
		return (ng.Status == "CREATE_COMPLETE"), nil
	}, 900*time.Second)
	th.AssertNoErr(t, err)

	t.Run("update", func(t *testing.T) { testNodeGroupUpdate(t, client, clusterID, nodeGroupID) })
	t.Run("delete", func(t *testing.T) { testNodeGroupDelete(t, client, clusterID, nodeGroupID) })
}

func testNodeGroupsList(t *testing.T, client *gophercloud.ServiceClient, clusterID string) {
	allPages, err := nodegroups.List(client, clusterID, nil).AllPages()
	th.AssertNoErr(t, err)

	allNodeGroups, err := nodegroups.ExtractNodeGroups(allPages)
	th.AssertNoErr(t, err)

	// By default two node groups should be created
	th.AssertEquals(t, 2, len(allNodeGroups))
}

func testNodeGroupGet(t *testing.T, client *gophercloud.ServiceClient, clusterID string) {
	listOpts := nodegroups.ListOpts{
		Role: "worker",
	}
	allPages, err := nodegroups.List(client, clusterID, listOpts).AllPages()
	th.AssertNoErr(t, err)

	allNodeGroups, err := nodegroups.ExtractNodeGroups(allPages)
	th.AssertNoErr(t, err)

	// Should be one worker node group
	th.AssertEquals(t, 1, len(allNodeGroups))

	ngID := allNodeGroups[0].UUID

	ng, err := nodegroups.Get(client, clusterID, ngID).Extract()
	th.AssertNoErr(t, err)

	// Should have got the same node group as from the list
	th.AssertEquals(t, ngID, ng.UUID)
	th.AssertEquals(t, "worker", ng.Role)
}

func testNodeGroupCreate(t *testing.T, client *gophercloud.ServiceClient, clusterID string) string {
	name := tools.RandomString("test-ng-", 8)

	// have to create two nodes for the Update test (can't set minimum above actual node count)
	two := 2
	createOpts := nodegroups.CreateOpts{
		Name:      name,
		NodeCount: &two,
	}

	ng, err := nodegroups.Create(client, clusterID, createOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, name, ng.Name)

	return ng.UUID
}

func testNodeGroupUpdate(t *testing.T, client *gophercloud.ServiceClient, clusterID, nodeGroupID string) {
	// Node group starts with min=1, max=unset
	// Set min, then set max, then set both

	updateOpts := []nodegroups.UpdateOptsBuilder{
		nodegroups.UpdateOpts{
			Op:    nodegroups.ReplaceOp,
			Path:  "/min_node_count",
			Value: 2,
		},
	}
	ng, err := nodegroups.Update(client, clusterID, nodeGroupID, updateOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, 2, ng.MinNodeCount)

	updateOpts = []nodegroups.UpdateOptsBuilder{
		nodegroups.UpdateOpts{
			Op:    nodegroups.ReplaceOp,
			Path:  "/max_node_count",
			Value: 5,
		},
	}
	ng, err = nodegroups.Update(client, clusterID, nodeGroupID, updateOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, false, ng.MaxNodeCount == nil)
	th.AssertEquals(t, 5, *ng.MaxNodeCount)

	updateOpts = []nodegroups.UpdateOptsBuilder{
		nodegroups.UpdateOpts{
			Op:    nodegroups.ReplaceOp,
			Path:  "/min_node_count",
			Value: 1,
		},
		nodegroups.UpdateOpts{
			Op:    nodegroups.ReplaceOp,
			Path:  "/max_node_count",
			Value: 3,
		},
	}
	ng, err = nodegroups.Update(client, clusterID, nodeGroupID, updateOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, false, ng.MaxNodeCount == nil)
	th.AssertEquals(t, 1, ng.MinNodeCount)
	th.AssertEquals(t, 3, *ng.MaxNodeCount)
}

func testNodeGroupDelete(t *testing.T, client *gophercloud.ServiceClient, clusterID, nodeGroupID string) {
	err := nodegroups.Delete(client, clusterID, nodeGroupID).ExtractErr()
	th.AssertNoErr(t, err)

	// Wait for the node group to be deleted
	err = tools.WaitFor(func() (bool, error) {
		_, err := nodegroups.Get(client, clusterID, nodeGroupID).Extract()
		if _, ok := err.(gophercloud.ErrDefault404); ok {
			return true, nil
		}
		return false, nil
	})
	th.AssertNoErr(t, err)
}
