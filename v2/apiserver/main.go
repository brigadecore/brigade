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
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
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
	var eventsStore core.EventsStore
	var jobsStore core.JobsStore
	var projectsStore core.ProjectsStore
	var secretsStore core.SecretsStore
	var serviceAccountsStore authx.ServiceAccountsStore
	var sessionsStore authx.SessionsStore
	var usersStore authx.UsersStore
	var workersStore core.WorkersStore
	{
		database, err := database(ctx)
		if err != nil {
			log.Fatal(err)
		}
		eventsStore, err = coreMongodb.NewEventsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		jobsStore, err = coreMongodb.NewJobsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		projectsStore, err = coreMongodb.NewProjectsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		secretsStore = coreKubernetes.NewSecretsStore(kubeClient)
		serviceAccountsStore, err = authxMongodb.NewServiceAccountsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		sessionsStore, err = authxMongodb.NewSessionsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		usersStore, err = authxMongodb.NewUsersStore(database)
		if err != nil {
			log.Fatal(err)
		}
		workersStore, err = coreMongodb.NewWorkersStore(database)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Message sending abstraction
	var queueWriterFactory queue.WriterFactory
	{
		config, err := writerFactoryConfig()
		if err != nil {
			log.Fatal(err)
		}
		queueWriterFactory = amqp.NewWriterFactory(config)
	}

	// Substrate
	var substrate core.Substrate
	{
		config, err := substrateConfig()
		if err != nil {
			log.Fatal(err)
		}
		substrate = coreKubernetes.NewSubstrate(
			kubeClient,
			queueWriterFactory,
			config,
		)
	}

	// Events service
	eventsService := core.NewEventsService(projectsStore, eventsStore, substrate)

	// Jobs service
	jobsService := core.NewJobsService(
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)

	// Projects service
	projectsService := core.NewProjectsService(projectsStore, substrate)

	// ServiceAccounts service
	serviceAccountsService :=
		authx.NewServiceAccountsService(serviceAccountsStore)

	// Secrets service
	secretsService := core.NewSecretsService(projectsStore, secretsStore)

	// Session service
	var sessionsService authx.SessionsService
	{
		config, err := sessionsServiceConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}
		sessionsService = authx.NewSessionsService(
			sessionsStore,
			usersStore,
			&config,
		)
	}

	// Substrate service
	substrateService := core.NewSubstrateService(substrate)

	// Users service
	usersService := authx.NewUsersService(usersStore, sessionsStore)

	// Workers service
	workersService := core.NewWorkersService(
		projectsStore,
		eventsStore,
		workersStore,
		substrate,
	)

	// Server
	var apiServer restmachinery.Server
	{
		authFilterConfig, err := tokenAuthFilterConfig(usersService.Get)
		if err != nil {
			log.Fatal(err)
		}
		authFilter := authn.NewTokenAuthFilter(
			serviceAccountsService.GetByToken,
			sessionsService.GetByToken,
			&authFilterConfig,
		)
		apiServerConfig, err := serverConfig()
		if err != nil {
			log.Fatal(err)
		}
		apiServer = restmachinery.NewServer(
			[]restmachinery.Endpoints{
				&coreREST.EventsEndpoints{
					AuthFilter: authFilter,
					EventSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/event.json",
					),
					Service: eventsService,
				},
				&coreREST.JobsEndpoints{
					AuthFilter: authFilter,
					JobStatusSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/job-status.json",
					),
					Service: jobsService,
				},
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
				&authxREST.ServiceAccountEndpoints{
					AuthFilter: authFilter,
					ServiceAccountSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/service-account.json",
					),
					Service: serviceAccountsService,
				},
				&authxREST.SessionsEndpoints{
					AuthFilter: authFilter,
					Service:    sessionsService,
				},
				&coreREST.SubstrateEndpoints{
					AuthFilter: authFilter,
					Service:    substrateService,
				},
				&authxREST.UsersEndpoints{
					AuthFilter: authFilter,
					Service:    usersService,
				},
				&coreREST.WorkersEndpoints{
					AuthFilter: authFilter,
					WorkerStatusSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/worker-status.json",
					),
					Service: workersService,
				},
			},
			&apiServerConfig,
		)
	}

	// Run it!
	log.Println(apiServer.ListenAndServe(ctx))
}
