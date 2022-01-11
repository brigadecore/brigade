package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/sdk/v2/authz"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
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
					Name: string(system.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						system.RoleEventCreator,
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
					Action: grantSystemRole(system.RoleEventCreator),
				},
				{
					Name: string(system.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of new projects.",
						system.RoleProjectCreator,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(system.RoleProjectCreator),
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
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List principals and their system-level roles",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:    flagRole,
					Aliases: []string{"r"},
					Usage:   "Narrow results to the specified role",
				},
				nonInteractiveFlag,
				&cli.StringFlag{
					Name:    flagServiceAccount,
					Aliases: []string{"s"},
					Usage: "Narrow results to the specified service account; " +
						"mutually exclusive with --user",
				},
				&cli.StringFlag{
					Name:    flagUser,
					Aliases: []string{"u"},
					Usage: "Narrow results to the specified user; mutually " +
						"exclusive with --service-account",
				},
				cliFlagOutput,
			},
			Action: listSystemRoles,
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
					Name: string(system.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						system.RoleEventCreator,
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
					Action: revokeSystemRole(system.RoleEventCreator),
				},
				{
					Name: string(system.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables creation of new projects.",
						system.RoleProjectCreator,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(system.RoleProjectCreator),
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
		if role == system.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(false)
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
			fmt.Printf(
				"Granted role %q to user %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Granted role %q to service account %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}

		return nil
	}
}

func listSystemRoles(c *cli.Context) error {
	userID := c.String(flagUser)
	serviceAccountID := c.String(flagServiceAccount)
	role := c.String(flagRole)
	output := c.String(flagOutput)

	if userID != "" && serviceAccountID != "" {
		return errors.New(
			"--user and --service-account filter flags are mutually exclusive",
		)
	}

	if err := validateOutputFormat(output); err != nil {
		return err
	}

	selector := authz.RoleAssignmentsSelector{
		Role: libAuthz.Role(role),
	}

	if userID != "" {
		selector.Principal = &libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeUser,
			ID:   userID,
		}
	} else if serviceAccountID != "" {
		selector.Principal = &libAuthz.PrincipalReference{
			Type: authz.PrincipalTypeServiceAccount,
			ID:   serviceAccountID,
		}
	}

	client, err := getClient(false)
	if err != nil {
		return err
	}

	opts := meta.ListOptions{}

	for {
		roleAssignments, err := client.Authz().RoleAssignments().List(
			c.Context,
			&selector,
			&opts,
		)
		if err != nil {
			return err
		}

		if len(roleAssignments.Items) == 0 {
			fmt.Println("No role assignments found.")
			return nil
		}

		switch strings.ToLower(output) {
		case flagOutputTable:
			table := uitable.New()
			table.AddRow("PRINCIPAL TYPE", "PRINCIPAL ID", "ROLE", "SCOPE")
			for _, roleAssignment := range roleAssignments.Items {
				table.AddRow(
					roleAssignment.Principal.Type,
					roleAssignment.Principal.ID,
					roleAssignment.Role,
					roleAssignment.Scope,
				)
			}
			fmt.Println(table)

		case flagOutputYAML:
			yamlBytes, err := yaml.Marshal(roleAssignments)
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from list role assignments operation",
				)
			}
			fmt.Println(string(yamlBytes))

		case flagOutputJSON:
			prettyJSON, err := json.MarshalIndent(roleAssignments, "", "  ")
			if err != nil {
				return errors.Wrap(
					err,
					"error formatting output from list role assignments operation",
				)
			}
			fmt.Println(string(prettyJSON))
		}

		if shouldContinue, err :=
			shouldContinue(
				c,
				roleAssignments.RemainingItemCount,
				roleAssignments.Continue,
			); err != nil {
			return err
		} else if !shouldContinue {
			break
		}

		opts.Continue = roleAssignments.Continue
	}

	return nil
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
		if role == system.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(false)
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
			fmt.Printf(
				"Revoked role %q for user %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}
		roleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Revoked role %q for service account %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}

		return nil
	}
}
