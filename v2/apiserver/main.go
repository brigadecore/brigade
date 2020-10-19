package main

// nolint: lll
import (
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	authxMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/authx/mongodb"
	authxREST "github.com/brigadecore/brigade/v2/apiserver/internal/authx/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	coreKubernetes "github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/core/mongodb"
	coreREST "github.com/brigadecore/brigade/v2/apiserver/internal/core/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery/authn"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
	"github.com/xeipuuv/gojsonschema"
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

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Data stores
	var projectsStore core.ProjectsStore
	var secretsStore core.SecretsStore
	var sessionsStore authx.SessionsStore
	var usersStore authx.UsersStore
	{
		database, err := mongodb.Database(ctx)
		if err != nil {
			log.Fatal(err)
		}
		projectsStore = coreMongodb.NewProjectsStore(database)
		secretsStore = coreKubernetes.NewSecretsStore(kubeClient)
		sessionsStore = authxMongodb.NewSessionsStore(database)
		usersStore = authxMongodb.NewUsersStore(database)
	}

	// Substrate
	substrate := coreKubernetes.NewSubstrate(kubeClient)

	// Projects service
	projectsService := core.NewProjectsService(
		projectsStore,
		substrate,
	)

	// Secrets service
	secretsService := core.NewSecretsService(projectsStore, secretsStore)

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
				&coreREST.ProjectsEndpoints{
					AuthFilter: authFilter,
					ProjectSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/project.json",
					),
					Service: projectsService,
				},
				&coreREST.SecretsEndpoints{
					AuthFilter: authFilter,
					SecretSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/secret.json",
					),
					Service: secretsService,
				},
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
