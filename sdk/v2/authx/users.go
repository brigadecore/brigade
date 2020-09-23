package authx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// User represents a (human) Brigade user.
type User struct {
	// ObjectMeta encapsulates User metadata.
	meta.ObjectMeta `json:"metadata"`
	// Name is the given name and surname of the User.
	Name string `json:"name,omitempty"`
	// Locked indicates when the User has been locked out of the system by an
	// administrator. If this field's value is nil, the User is not locked.
	Locked *time.Time `json:"locked,omitempty"`
	// Roles is a slice of Roles assigned to this User.
	Roles []Role `json:"roles,omitempty"`
}

// MarshalJSON amends User instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "User",
			},
			Alias: (Alias)(u),
		},
	)
}

// UserList is an ordered and pageable list of Users.
type UserList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Users.
	Items []User `json:"items,omitempty"`
}

// MarshalJSON amends UserList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
func (u UserList) MarshalJSON() ([]byte, error) {
	type Alias UserList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "UserList",
			},
			Alias: (Alias)(u),
		},
	)
}

// UsersSelector represents useful filter criteria when selecting multiple Users
// for API group operations like list. It currently has no fields, but exists to
// preserve the possibility of future expansion without having to change client
// function signatures.
type UsersSelector struct{}

// UsersClient is the specialized client for managing Users with the Brigade
// API.
type UsersClient interface {
	// List returns a UserList.
	List(context.Context, *UsersSelector, *meta.ListOptions) (UserList, error)
	// Get retrieves a single User specified by their identifier.
	Get(context.Context, string) (User, error)

	// Lock revokes system access for a single User specified by their identifier.
	// Implementations MUST also delete or invalidate any and all of the User's
	// existing Sessions.
	Lock(context.Context, string) error
	// Unlock restores system access for a single User (after presumably having
	// been revoked) specified by their identifier.
	Unlock(context.Context, string) error
}

type usersClient struct {
	*rm.BaseClient
}

// NewUsersClient returns a specialized client for managing Users.
func NewUsersClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) UsersClient {
	return &usersClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (u *usersClient) List(
	ctx context.Context,
	_ *UsersSelector,
	opts *meta.ListOptions,
) (UserList, error) {
	users := UserList{}
	return users, u.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/users",
			AuthHeaders: u.BearerTokenAuthHeaders(),
			QueryParams: u.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &users,
		},
	)
}

func (u *usersClient) Get(ctx context.Context, id string) (User, error) {
	user := User{}
	return user, u.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/users/%s", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &user,
		},
	)
}

func (u *usersClient) Lock(ctx context.Context, id string) error {
	return u.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (u *usersClient) Unlock(ctx context.Context, id string) error {
	return u.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/users/%s/lock", id),
			AuthHeaders: u.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
