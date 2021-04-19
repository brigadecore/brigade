package main

import (
	"fmt"

	"github.com/brigadecore/brigade/sdk/v2/authz"
	"github.com/brigadecore/brigade/sdk/v2/core"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var (
	roleGrantFlags = []cli.Flag{
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
	roleRevokeFlags = []cli.Flag{
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

var rolesCommands = &cli.Command{
	Name:    "role",
	Aliases: []string{"roles"},
	Usage:   "Manage system roles for users or service accounts",
	Subcommands: []*cli.Command{
		{
			Name:  "grant",
			Usage: "Grant a system-level role to a user or service account",
			Subcommands: []*cli.Command{
				{
					Name: string(system.RoleAdmin),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables system management including "+
							"system-level permissions for other users and service accounts.",
						system.RoleAdmin,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(system.RoleAdmin),
				},
				{
					Name: string(core.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						core.RoleEventCreator,
					),
					Flags: append(
						roleGrantFlags,
						&cli.StringFlag{ // Special flag for EVENT_CREATOR
							Name: flagSource,
							Usage: "Permit creation of events from the specified " +
								"source only (required)",
							Required: true,
						},
					),
					Action: grantSystemRole(core.RoleEventCreator),
				},
				{
					Name: string(core.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of new projects.",
						core.RoleProjectCreator,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(core.RoleProjectCreator),
				},
				{
					Name: string(system.RoleReader),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables global read-only access to "+
							"Brigade.",
						system.RoleReader,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(system.RoleReader),
				},
			},
		},
		{
			Name:  "revoke",
			Usage: "Revoke a system-level role from a user or service account",
			Subcommands: []*cli.Command{
				{
					Name: string(system.RoleAdmin),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables system management including "+
							"system-level permissions for other users and service accounts.",
						system.RoleAdmin,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(system.RoleAdmin),
				},
				{
					Name: string(core.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						core.RoleEventCreator,
					),
					Flags: append(
						roleRevokeFlags,
						&cli.StringFlag{ // Special flag for EVENT_CREATOR
							Name: flagSource,
							Usage: "Revoke creation of events from the specified " +
								"source only (required)",
							Required: true,
						},
					),
					Action: revokeSystemRole(core.RoleEventCreator),
				},
				{
					Name: string(core.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables creation of new projects.",
						core.RoleProjectCreator,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(core.RoleProjectCreator),
				},
				{
					Name: string(system.RoleReader),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables global read-only access to "+
							"Brigade.",
						system.RoleReader,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(system.RoleReader),
				},
			},
		},
	},
}

func grantSystemRole(role libAuthz.Role) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		roleAssignment := libAuthz.RoleAssignment{
			Role: role,
		}

		// Special logic for EVENT_CREATOR
		if role == core.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(c)
		if err != nil {
			return err
		}

		roleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}

		return nil
	}
}

func revokeSystemRole(role libAuthz.Role) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		roleAssignment := libAuthz.RoleAssignment{
			Role: role,
		}

		// Special logic for EVENT_CREATOR
		if role == core.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(c)
		if err != nil {
			return err
		}

		roleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
		}

		return nil
	}
}
