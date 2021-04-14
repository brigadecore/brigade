package main

import (
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/authz"
	"github.com/brigadecore/brigade/sdk/v2/core"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	projectRoleGrantFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     flagID,
			Aliases:  []string{"i", flagProject, "p"},
			Usage:    "Grant the role for the specified project (required)",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:    flagUser,
			Aliases: []string{"u"},
			Usage:   "Grant the role to the specified user",
		},
		&cli.StringSliceFlag{
			Name:    flagServiceAccount,
			Aliases: []string{"s"},
			Usage:   "Grant the role to the specified service account",
		},
	}
	projectRoleRevokeFlags = []cli.Flag{
		&cli.StringFlag{
			Name:     flagID,
			Aliases:  []string{"i", flagProject, "p"},
			Usage:    "Revoke the role for the specified project (required)",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:    flagUser,
			Aliases: []string{"u"},
			Usage:   "Revoke the role for the specified user",
		},
		&cli.StringSliceFlag{
			Name:    flagServiceAccount,
			Aliases: []string{"s"},
			Usage:   "Revoke the role for the specified service account",
		},
	}
)

var projectRolesCommands = &cli.Command{
	Name:    "role",
	Aliases: []string{"roles"},
	Usage:   "Manage project roles",
	Subcommands: []*cli.Command{
		{
			Name:  "grant",
			Usage: "Grant a project-level role to a user or service account",
			Subcommands: []*cli.Command{
				{
					Name: string(core.RoleNameProjectAdmin),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables management of all "+
							"aspects of the project, including its secrets, as well as "+
							"project-level permissions for other users and service "+
							"accounts.",
						core.RoleNameProjectAdmin,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleNameProjectAdmin),
				},
				{
					Name: string(core.RoleNameProjectDeveloper),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables updating the project "+
							"definition, but does NOT enable management of the project's "+
							"secrets or project-level permissions for other users and "+
							"service accounts.",
						core.RoleNameProjectDeveloper,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleNameProjectDeveloper),
				},
				{
					Name: string(core.RoleNameProjectUser),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables creation and management "+
							"of events associated with the project",
						core.RoleNameProjectUser,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleNameProjectUser),
				},
			},
		},
		{
			Name:  "revoke",
			Usage: "Revoke a project-level role from a user or service account",
			Subcommands: []*cli.Command{
				{
					Name: string(core.RoleNameProjectAdmin),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables management of all "+
							"aspects of the project, including its secrets, as well as "+
							"project-level permissions for other users and service "+
							"accounts.",
						core.RoleNameProjectAdmin,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleNameProjectAdmin),
				},
				{
					Name: string(core.RoleNameProjectDeveloper),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables updating the project "+
							"definition, but does NOT enable management of the project's "+
							"secrets or project-level permissions for other users and "+
							"service accounts.",
						core.RoleNameProjectDeveloper,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleNameProjectDeveloper),
				},
				{
					Name: string(core.RoleNameProjectUser),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables creation and "+
							"management of events associated with the project",
						core.RoleNameProjectUser,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleNameProjectUser),
				},
			},
		},
	},
}

func grantProjectRole(
	roleName libAuthz.RoleName,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		projectID := c.String(flagID)
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		client, err := getClient(c)
		if err != nil {
			return err
		}

		roleAssignment := libAuthz.RoleAssignment{
			Role: libAuthz.Role{
				Type:  core.RoleTypeProject,
				Name:  roleName,
				Scope: projectID,
			},
		}

		roleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}

		return nil
	}
}

func revokeProjectRole(
	roleName libAuthz.RoleName,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		projectID := c.String(flagID)
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		client, err := getClient(c)
		if err != nil {
			return err
		}

		roleAssignment := libAuthz.RoleAssignment{
			Role: libAuthz.Role{
				Type:  core.RoleTypeProject,
				Name:  roleName,
				Scope: projectID,
			},
		}

		roleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}

		return nil
	}
}
