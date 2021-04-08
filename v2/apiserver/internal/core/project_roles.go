package core

import libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"

// ProjectIDGlobal represents an unbounded scope.
const ProjectIDGlobal = "*"

// ProjectRole represents a set of project-level permissions.
type ProjectRole struct {
	// Name is the name of a ProjectRole and has domain-specific meaning.
	Name libAuthz.RoleName `json:"name,omitempty" bson:"name,omitempty"`
	// ProjectID qualifies the scope of the ProjectRole.
	ProjectID string `json:"projectID,omitempty" bson:"projectID,omitempty"`
}

// Matches determines if this ProjectRole matches the requiredRole argument.
// This ProjectRole is a match for the required one if the Name fields have the
// same values in both AND if the value of this ProjectRole's ProjectID field is
// either the same as that of the required ProjectRole's ProjectID field OR is
// unbounded ("*").
//
// Note that order is important. A.Matches(B) does not guarantee B.Matches(A).
func (p ProjectRole) Matches(requiredRole ProjectRole) bool {
	return p.Name == requiredRole.Name &&
		(p.ProjectID == requiredRole.ProjectID || p.ProjectID == ProjectIDGlobal)
}
