package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/brigadecore/brigade/sdk/v3/authz"
	"github.com/brigadecore/brigade/sdk/v3/core"
	libAuthz "github.com/brigadecore/brigade/sdk/v3/lib/authz"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
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
					Name: string(core.RoleProjectAdmin),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables management of all "+
							"aspects of the project, including its secrets, as well as "+
							"project-level permissions for other users and service "+
							"accounts.",
						core.RoleProjectAdmin,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleProjectAdmin),
				},
				{
					Name: string(core.RoleProjectDeveloper),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables updating the project "+
							"definition, but does NOT enable management of the project's "+
							"secrets or project-level permissions for other users and "+
							"service accounts.",
						core.RoleProjectDeveloper,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleProjectDeveloper),
				},
				{
					Name: string(core.RoleProjectUser),
					Usage: fmt.Sprintf(
						"Grant the %s project role, which enables creation and management "+
							"of events associated with the project",
						core.RoleProjectUser,
					),
					Flags:  projectRoleGrantFlags,
					Action: grantProjectRole(core.RoleProjectUser),
				},
			},
		},
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "List principals and their project-level roles",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     flagID,
					Aliases:  []string{"i", flagProject, "p"},
					Usage:    "List principals and their roles for the specified project",
					Required: true,
				},
				nonInteractiveFlag,
				&cli.StringFlag{
					Name:    flagRole,
					Aliases: []string{"r"},
					Usage:   "Narrow results to the specified role",
				},
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
			Action: listProjectRoles,
		},
		{
			Name:  "revoke",
			Usage: "Revoke a project-level role from a user or service account",
			Subcommands: []*cli.Command{
				{
					Name: string(core.RoleProjectAdmin),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables management of all "+
							"aspects of the project, including its secrets, as well as "+
							"project-level permissions for other users and service "+
							"accounts.",
						core.RoleProjectAdmin,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleProjectAdmin),
				},
				{
					Name: string(core.RoleProjectDeveloper),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables updating the project "+
							"definition, but does NOT enable management of the project's "+
							"secrets or project-level permissions for other users and "+
							"service accounts.",
						core.RoleProjectDeveloper,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleProjectDeveloper),
				},
				{
					Name: string(core.RoleProjectUser),
					Usage: fmt.Sprintf(
						"Revoke the %s project role, which enables creation and "+
							"management of events associated with the project",
						core.RoleProjectUser,
					),
					Flags:  projectRoleRevokeFlags,
					Action: revokeProjectRole(core.RoleProjectUser),
				},
			},
		},
	},
}

func grantProjectRole(role libAuthz.Role) func(c *cli.Context) error {
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

		client, err := getClient(false)
		if err != nil {
			return err
		}

		projectRoleAssignment := core.ProjectRoleAssignment{
			Role: role,
		}

		projectRoleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, projectRoleAssignment.Principal.ID = range userIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Grant(
				c.Context,
				projectID,
				projectRoleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Granted role %q for project %q to user %q.\n",
				projectRoleAssignment.Role,
				projectID,
				projectRoleAssignment.Principal.ID,
			)
		}
		projectRoleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, projectRoleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Grant(
				c.Context,
				projectID,
				projectRoleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Granted role %q for project %q to service account %q.\n",
				projectRoleAssignment.Role,
				projectID,
				projectRoleAssignment.Principal.ID,
			)
		}

		return nil
	}
}

func listProjectRoles(c *cli.Context) error {
	projectID := c.String(flagProject)
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

	selector := core.ProjectRoleAssignmentsSelector{
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
		roleAssignments, err :=
			client.Core().Projects().Authz().RoleAssignments().List(
				c.Context,
				projectID,
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
			table.AddRow("PRINCIPAL TYPE", "PRINCIPAL ID", "ROLE")
			for _, roleAssignment := range roleAssignments.Items {
				table.AddRow(
					roleAssignment.Principal.Type,
					roleAssignment.Principal.ID,
					roleAssignment.Role,
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

func revokeProjectRole(role libAuthz.Role) func(c *cli.Context) error {
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

		client, err := getClient(false)
		if err != nil {
			return err
		}

		projectRoleAssignment := core.ProjectRoleAssignment{
			Role: role,
		}

		projectRoleAssignment.Principal.Type = authz.PrincipalTypeUser
		for _, projectRoleAssignment.Principal.ID = range userIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Revoke(
				c.Context,
				projectID,
				projectRoleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Revoked role %q for project %q from user %q.\n",
				projectRoleAssignment.Role,
				projectID,
				projectRoleAssignment.Principal.ID,
			)
		}
		projectRoleAssignment.Principal.Type = authz.PrincipalTypeServiceAccount
		for _, projectRoleAssignment.Principal.ID = range serviceAccountIDs {
			if err = client.Core().Projects().Authz().RoleAssignments().Revoke(
				c.Context,
				projectID,
				projectRoleAssignment,
				nil,
			); err != nil {
				return err
			}
			fmt.Printf(
				"Revoked role %q for project %q from service account %q.\n",
				projectRoleAssignment.Role,
				projectID,
				projectRoleAssignment.Principal.ID,
			)
		}

		return nil
	}
}
