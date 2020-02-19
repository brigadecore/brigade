// +build acceptance identity trusts

package v3

import (
	"testing"

	"github.com/gophercloud/gophercloud/acceptance/clients"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/trusts"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/roles"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/tokens"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/users"
	th "github.com/gophercloud/gophercloud/testhelper"
)

func TestTrustCRUD(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewIdentityV3Client()
	th.AssertNoErr(t, err)

	// Generate a token and obtain the Admin user's ID from it.
	ao, err := openstack.AuthOptionsFromEnv()
	th.AssertNoErr(t, err)

	authOptions := tokens.AuthOptions{
		Username:   ao.Username,
		Password:   ao.Password,
		DomainName: ao.DomainName,
		DomainID:   ao.DomainID,
	}

	token, err := tokens.Create(client, &authOptions).Extract()
	th.AssertNoErr(t, err)
	adminUser, err := tokens.Get(client, token.ID).ExtractUser()
	th.AssertNoErr(t, err)

	// Get the admin and member role IDs.
	adminRoleID := ""
	memberRoleID := ""
	allPages, err := roles.List(client, nil).AllPages()
	th.AssertNoErr(t, err)
	allRoles, err := roles.ExtractRoles(allPages)
	th.AssertNoErr(t, err)

	for _, v := range allRoles {
		if v.Name == "admin" {
			adminRoleID = v.ID
		}

		if v.Name == "member" {
			memberRoleID = v.ID
		}
	}

	// Create a project to apply the trust.
	trusteeProject, err := CreateProject(t, client, nil)
	th.AssertNoErr(t, err)
	defer DeleteProject(t, client, trusteeProject.ID)

	tools.PrintResource(t, trusteeProject)

	// Add the admin user to the trustee project.
	assignOpts := roles.AssignOpts{
		UserID:    adminUser.ID,
		ProjectID: trusteeProject.ID,
	}

	err = roles.Assign(client, adminRoleID, assignOpts).ExtractErr()
	th.AssertNoErr(t, err)

	// Create a user as the trustee.
	trusteeUserCreateOpts := users.CreateOpts{
		Password: "secret",
		DomainID: "default",
	}
	trusteeUser, err := CreateUser(t, client, &trusteeUserCreateOpts)
	th.AssertNoErr(t, err)
	defer DeleteUser(t, client, trusteeUser.ID)

	// Create a trust.
	trust, err := CreateTrust(t, client, trusts.CreateOpts{
		TrusteeUserID: trusteeUser.ID,
		TrustorUserID: adminUser.ID,
		ProjectID:     trusteeProject.ID,
		Roles: []trusts.Role{
			{
				ID: memberRoleID,
			},
		},
	})
	th.AssertNoErr(t, err)
	defer DeleteTrust(t, client, trust.ID)
}
