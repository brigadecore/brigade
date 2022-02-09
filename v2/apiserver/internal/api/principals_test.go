package api

import (
	"context"
	"testing"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestContextWithPrincipal(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	ctx := ContextWithPrincipal(context.Background(), &testUser)
	val := ctx.Value(principalContextKey{})
	require.Equal(t, &testUser, val)
}

func TestPrincipalFromContext(t *testing.T) {
	testUser := User{
		ObjectMeta: meta.ObjectMeta{
			ID: "tony@starkindustries.com",
		},
	}
	ctx :=
		context.WithValue(context.Background(), principalContextKey{}, &testUser)
	principal := PrincipalFromContext(ctx)
	require.Equal(t, &testUser, principal)
}

func TestNewPrincipalsService(t *testing.T) {
	svc, ok := NewPrincipalsService(alwaysAuthorize).(*principalsService)
	require.True(t, ok)
	require.NotNil(t, svc.authorize)
}

func TestPrincipalsServiceWhoAmI(t *testing.T) {
	const testID = "gozer"
	testCases := []struct {
		name       string
		ctx        context.Context
		service    PrincipalsService
		assertions func(PrincipalReference, error)
	}{
		{
			name: "unauthorized",
			ctx:  context.Background(),
			service: &principalsService{
				authorize: neverAuthorize,
			},
			assertions: func(_ PrincipalReference, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "principal is root",
			ctx: ContextWithPrincipal(
				context.Background(),
				&RootPrincipal{},
			),
			service: &principalsService{
				authorize: alwaysAuthorize,
			},
			assertions: func(ref PrincipalReference, err error) {
				require.NoError(t, err)
				require.Equal(t, PrincipalTypeRoot, ref.Type)
				require.Equal(t, "root", ref.ID)
			},
		},
		{
			name: "principal is a service account",
			ctx: ContextWithPrincipal(
				context.Background(),
				&ServiceAccount{
					ObjectMeta: meta.ObjectMeta{
						ID: testID,
					},
				},
			),
			service: &principalsService{
				authorize: alwaysAuthorize,
			},
			assertions: func(ref PrincipalReference, err error) {
				require.NoError(t, err)
				require.Equal(t, PrincipalTypeServiceAccount, ref.Type)
				require.Equal(t, testID, ref.ID)
			},
		},
		{
			name: "principal is a user",
			ctx: ContextWithPrincipal(
				context.Background(),
				&User{
					ObjectMeta: meta.ObjectMeta{
						ID: testID,
					},
				},
			),
			service: &principalsService{
				authorize: alwaysAuthorize,
			},
			assertions: func(ref PrincipalReference, err error) {
				require.NoError(t, err)
				require.Equal(t, PrincipalTypeUser, ref.Type)
				require.Equal(t, testID, ref.ID)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ref, err := testCase.service.WhoAmI(testCase.ctx)
			testCase.assertions(ref, err)
		})
	}
}
