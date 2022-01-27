package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
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
					Name: string(sdk.RoleAdmin),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables system management including "+
							"system-level permissions for other users and service accounts.",
						sdk.RoleAdmin,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(sdk.RoleAdmin),
				},
				{
					Name: string(sdk.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						sdk.RoleEventCreator,
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
					Action: grantSystemRole(sdk.RoleEventCreator),
				},
				{
					Name: string(sdk.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of new projects.",
						sdk.RoleProjectCreator,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(sdk.RoleProjectCreator),
				},
				{
					Name: string(sdk.RoleReader),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables global read-only access to "+
							"Brigade.",
						sdk.RoleReader,
					),
					Flags:  roleGrantFlags,
					Action: grantSystemRole(sdk.RoleReader),
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
					Name: string(sdk.RoleAdmin),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables system management including "+
							"system-level permissions for other users and service accounts.",
						sdk.RoleAdmin,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(sdk.RoleAdmin),
				},
				{
					Name: string(sdk.RoleEventCreator),
					Usage: fmt.Sprintf(
						"Grant the %s role, which enables creation of events for all "+
							"projects.",
						sdk.RoleEventCreator,
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
					Action: revokeSystemRole(sdk.RoleEventCreator),
				},
				{
					Name: string(sdk.RoleProjectCreator),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables creation of new projects.",
						sdk.RoleProjectCreator,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(sdk.RoleProjectCreator),
				},
				{
					Name: string(sdk.RoleReader),
					Usage: fmt.Sprintf(
						"Revoke the %s role, which enables global read-only access to "+
							"Brigade.",
						sdk.RoleReader,
					),
					Flags:  roleRevokeFlags,
					Action: revokeSystemRole(sdk.RoleReader),
				},
			},
		},
	},
}

func grantSystemRole(role sdk.Role) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		roleAssignment := sdk.RoleAssignment{
			Role: role,
		}

		// Special logic for EVENT_CREATOR
		if role == sdk.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(false)
		if err != nil {
			return err
		}

		roleAssignment.Principal.Type = sdk.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Granted role %q to user %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}
		roleAssignment.Principal.Type = sdk.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Grant(
				c.Context,
				roleAssignment,
				nil,
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

	selector := sdk.RoleAssignmentsSelector{
		Role: sdk.Role(role),
	}

	if userID != "" {
		selector.Principal = &sdk.PrincipalReference{
			Type: sdk.PrincipalTypeUser,
			ID:   userID,
		}
	} else if serviceAccountID != "" {
		selector.Principal = &sdk.PrincipalReference{
			Type: sdk.PrincipalTypeServiceAccount,
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

func revokeSystemRole(role sdk.Role) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		userIDs := c.StringSlice(flagUser)
		serviceAccountIDs := c.StringSlice(flagServiceAccount)
		if len(userIDs) == 0 && len(serviceAccountIDs) == 0 {
			return errors.New(
				"at least one user or service account must be specified using the " +
					"--user or --service-account flags",
			)
		}

		roleAssignment := sdk.RoleAssignment{
			Role: role,
		}

		// Special logic for EVENT_CREATOR
		if role == sdk.RoleEventCreator {
			roleAssignment.Scope = c.String(flagSource)
		}

		client, err := getClient(false)
		if err != nil {
			return err
		}

		roleAssignment.Principal.Type = sdk.PrincipalTypeUser
		for _, roleAssignment.Principal.ID = range userIDs {
			if err = client.Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Revoked role %q for user %q.\n",
				roleAssignment.Role,
				roleAssignment.Principal.ID,
			)
		}
		roleAssignment.Principal.Type = sdk.PrincipalTypeServiceAccount
		for _, roleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Authz().RoleAssignments().Revoke(
				c.Context,
				roleAssignment,
				nil,
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
