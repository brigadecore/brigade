package main

import (
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	authxMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/authx/mongodb" // nolint: lll
	authxREST "github.com/brigadecore/brigade/v2/apiserver/internal/authx/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
)

// main wires up the dependency graph for the API server, then runs the API
// server unit it is interrupted or experiences a fatal error.
func main() {
	log.Printf(
		"Starting Brigade API Server -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	ctx := signals.Context()

	// Data stores
	var sessionsStore authx.SessionsStore
	var usersStore authx.UsersStore
	{
		database, err := mongodb.Database(ctx)
		if err != nil {
			log.Fatal(err)
		}
		sessionsStore = authxMongodb.NewSessionsStore(database)
		usersStore = authxMongodb.NewUsersStore(database)
	}

	// Session service
	var sessionsService authx.SessionsService
	{
		config, err := authx.GetSessionsServiceConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}
		sessionsService = authx.NewSessionsService(
			sessionsStore,
			usersStore,
			&config,
		)
	}

	// Server
	var apiServer restmachinery.Server
	{
		// TODO: Once the UsersService is implemented, replace the store function
		// below with the service function.
		authFilterConfig, err := authn.GetTokenAuthFilterConfig(usersStore.Get)
		if err != nil {
			log.Fatal(err)
		}
		authFilter := authn.NewTokenAuthFilter(
			sessionsService.GetByToken,
			&authFilterConfig,
		)
		apiServerConfig, err := restmachinery.GetServerConfig()
		if err != nil {
			log.Fatal(err)
		}
		apiServer = restmachinery.NewServer(
			[]restmachinery.Endpoints{
				&authxREST.SessionsEndpoints{
					AuthFilter: authFilter,
					Service:    sessionsService,
				},
			},
			&apiServerConfig,
		)
	}

	// Run it!
	log.Println(apiServer.ListenAndServe(ctx))
}
