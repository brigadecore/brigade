package core

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	"github.com/pkg/errors"
)

type projectRoleAssignmentsService struct {
	projectsStore        ProjectsStore
	usersStore           authn.UsersStore
	serviceAccountsStore authn.ServiceAccountsStore
	roleAssignmentsStore authz.RoleAssignmentsStore
}

// NewProjectRoleAssignmentsService returns a specialized interface for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsService(
	projectsStore ProjectsStore,
	usersStore authn.UsersStore,
	serviceAccountsStore authn.ServiceAccountsStore,
	roleAssignmentsStore authz.RoleAssignmentsStore,
) authz.RoleAssignmentsService {
	return &projectRoleAssignmentsService{
		projectsStore:        projectsStore,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		roleAssignmentsStore: roleAssignmentsStore,
	}
}

func (p *projectRoleAssignmentsService) Grant(
	ctx context.Context,
	roleAssignment authz.RoleAssignment,
) error {
	projectID := roleAssignment.Role.Scope

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	if roleAssignment.Principal.Type == authz.PrincipalTypeUser {
		// Make sure the User exists
		if _, err :=
			p.usersStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
	} else if roleAssignment.Principal.Type == authz.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			p.serviceAccountsStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.Principal.ID,
			)
		}
	} else {
		return nil
	}

	// Give them the Role
	if err := p.roleAssignmentsStore.Grant(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error granting project %q role %q to %s %q in store",
			projectID,
			roleAssignment.Role.Name,
			roleAssignment.Principal.Type,
			roleAssignment.Principal.ID,
		)
	}

	return nil
}

func (p *projectRoleAssignmentsService) Revoke(
	ctx context.Context,
	roleAssignment authz.RoleAssignment,
) error {
	projectID := roleAssignment.Role.Scope

	// Make sure the project exists
	_, err := p.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}

	if roleAssignment.Principal.Type == authz.PrincipalTypeUser {
		// Make sure the User exists
		if _, err :=
			p.usersStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving user %q from store",
				roleAssignment.Principal.ID,
			)
		}
	} else if roleAssignment.Principal.Type == authz.PrincipalTypeServiceAccount {
		// Make sure the ServiceAccount exists
		if _, err :=
			p.serviceAccountsStore.Get(ctx, roleAssignment.Principal.ID); err != nil {
			return errors.Wrapf(
				err,
				"error retrieving service account %q from store",
				roleAssignment.Principal.ID,
			)
		}
	} else {
		return nil
	}

	// Revoke the Role
	if err := p.roleAssignmentsStore.Revoke(ctx, roleAssignment); err != nil {
		return errors.Wrapf(
			err,
			"error revoking project %q role %q for %s %q in store",
			projectID,
			roleAssignment.Role.Name,
			roleAssignment.Principal.Type,
			roleAssignment.Principal.ID,
		)
	}
	return nil
}
