package main

// nolint: lll
import (
	"context"
	"log"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade-foundations/signals"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	apiKubernetes "github.com/brigadecore/brigade/v2/apiserver/internal/api/kubernetes"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/mongodb"
	"github.com/brigadecore/brigade/v2/apiserver/internal/api/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/assets"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/xeipuuv/gojsonschema"
	"go.mongodb.org/mongo-driver/mongo"
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
	var database *mongo.Database
	{
		if database, err = databaseConnection(ctx); err != nil {
			log.Fatal(err)
		}
		// No network I/O occurs when creating the DB connection, so we'll test it
		// here. This will block until the connection is verified or max retries are
		// exhausted. What we're trying to prevent is both 1. moving on in the
		// startup process without the database available AND 2. crashing too
		// prematurely while waiting for the database become available.
		if err = testDatabaseConnection(ctx, database); err != nil {
			log.Fatal(err)
		}
	}

	var coolLogsStore api.CoolLogsStore
	var eventsStore api.EventsStore
	var jobsStore api.JobsStore
	var projectsStore api.ProjectsStore
	var projectRoleAssignmentsStore api.ProjectRoleAssignmentsStore
	var roleAssignmentsStore api.RoleAssignmentsStore
	var secretsStore api.SecretsStore
	var serviceAccountsStore api.ServiceAccountsStore
	var sessionsStore api.SessionsStore
	var usersStore api.UsersStore
	var warmLogsStore api.LogsStore
	var workersStore api.WorkersStore
	{
		coolLogsStore = mongodb.NewLogsStore(database)
		eventsStore, err = mongodb.NewEventsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		jobsStore, err = mongodb.NewJobsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		projectsStore, err = mongodb.NewProjectsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		projectRoleAssignmentsStore =
			mongodb.NewProjectRoleAssignmentsStore(database)
		roleAssignmentsStore = mongodb.NewRoleAssignmentsStore(database)
		secretsStore = apiKubernetes.NewSecretsStore(kubeClient)
		serviceAccountsStore, err = mongodb.NewServiceAccountsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		sessionsStore, err = mongodb.NewSessionsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		usersStore, err = mongodb.NewUsersStore(database)
		if err != nil {
			log.Fatal(err)
		}
		warmLogsStore = apiKubernetes.NewLogsStore(kubeClient)
		workersStore, err = mongodb.NewWorkersStore(database)
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
		// No network I/O occurs when constructing a queue.WriterFactory, so we'll
		// test the underlying connection here by using the factory to get a
		// queue.Writer for the healthz queue. This will block until the underlying
		// connection succeeds or max retries are exhausted. What we're trying to
		// prevent is BOTH 1. moving on in the startup process without the message
		// server available AND 2. crashing too prematurely while waiting for the
		// message server to become available.
		if _, err := queueWriterFactory.NewWriter("healthz"); err != nil {
			log.Fatal(err)
		}
	}

	// Substrate
	var substrate api.Substrate
	{
		config, err := substrateConfig()
		if err != nil {
			log.Fatal(err)
		}
		substrate = apiKubernetes.NewSubstrate(
			kubeClient,
			queueWriterFactory,
			config,
		)
	}

	// Authorizers
	authorizer := api.NewAuthorizer(roleAssignmentsStore)
	projectAuthorizer := api.NewProjectAuthorizer(projectRoleAssignmentsStore)

	// Events service
	eventsService := api.NewEventsService(
		authorizer.Authorize,
		projectAuthorizer.Authorize,
		projectsStore,
		eventsStore,
		coolLogsStore,
		substrate,
	)

	// Jobs service
	jobsService := api.NewJobsService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)

	// Logs service
	logsService := api.NewLogsService(
		authorizer.Authorize,
		projectAuthorizer.Authorize,
		projectsStore,
		eventsStore,
		warmLogsStore,
		coolLogsStore,
	)

	// Principals service
	principalsService := api.NewPrincipalsService(authorizer.Authorize)

	// Projects service
	projectsService := api.NewProjectsService(
		authorizer.Authorize,
		projectAuthorizer.Authorize,
		projectsStore,
		eventsStore,
		coolLogsStore,
		projectRoleAssignmentsStore,
		substrate,
	)

	// ProjectRoleAssignments service
	projectRoleAssignmentsService := api.NewProjectRoleAssignmentsService(
		authorizer.Authorize,
		projectAuthorizer.Authorize,
		projectsStore,
		usersStore,
		serviceAccountsStore,
		projectRoleAssignmentsStore,
	)

	// Roles service
	roleAssignmentsService := api.NewRoleAssignmentsService(
		authorizer.Authorize,
		usersStore,
		serviceAccountsStore,
		roleAssignmentsStore,
	)

	// ServiceAccounts service
	serviceAccountsService := api.NewServiceAccountsService(
		authorizer.Authorize,
		serviceAccountsStore,
		roleAssignmentsStore,
		projectRoleAssignmentsStore,
	)

	// Secrets service
	secretsService := api.NewSecretsService(
		authorizer.Authorize,
		projectAuthorizer.Authorize,
		projectsStore,
		secretsStore,
	)

	// Session service
	var sessionsService api.SessionsService
	{
		thirdPartyAuthHelper, err := thirdPartyAuthHelper(ctx)
		if err != nil {
			log.Fatal(err)
		}
		config, err := sessionsServiceConfig()
		if err != nil {
			log.Fatal(err)
		}
		sessionsService = api.NewSessionsService(
			sessionsStore,
			usersStore,
			roleAssignmentsStore,
			thirdPartyAuthHelper,
			&config,
		)
	}

	// Substrate service
	substrateService := api.NewSubstrateService(authorizer.Authorize, substrate)

	// Users service
	usersService := api.NewUsersService(
		authorizer.Authorize,
		usersStore,
		sessionsStore,
		roleAssignmentsStore,
		projectRoleAssignmentsStore,
		usersServiceConfig(),
	)

	// Workers service
	workersService := api.NewWorkersService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		workersStore,
		substrate,
	)

	// Server
	var apiServer restmachinery.Server
	{
		authFilterConfig, err := tokenAuthFilterConfig(usersStore.Get)
		if err != nil {
			log.Fatal(err)
		}
		authFilter := rest.NewTokenAuthFilter(
			serviceAccountsService.GetByToken,
			sessionsService.GetByToken,
			eventsService.GetByWorkerToken,
			&authFilterConfig,
		)
		apiServerConfig, err := serverConfig()
		if err != nil {
			log.Fatal(err)
		}
		apiServer = restmachinery.NewServer(
			[]restmachinery.Endpoints{
				&rest.AuthnEndpoints{
					AuthFilter: authFilter,
					Service:    principalsService,
				},
				&rest.EventsEndpoints{
					AuthFilter: authFilter,
					EventSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/event.json",
					),
					SourceStateSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/source-state.json",
					),
					EventSummarySchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/event-summary.json",
					),
					Service: eventsService,
				},
				&rest.JobsEndpoints{
					AuthFilter: authFilter,
					JobSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/job.json",
					),
					JobStatusSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/job-status.json",
					),
					Service: jobsService,
				},
				&rest.LogsEndpoints{
					AuthFilter: authFilter,
					Service:    logsService,
				},
				&rest.ProjectsEndpoints{
					AuthFilter: authFilter,
					ProjectSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/project.json",
					),
					Service: projectsService,
				},
				&rest.ProjectRoleAssignmentsEndpoints{
					AuthFilter: authFilter,
					ProjectRoleAssignmentSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/project-role-assignment.json",
					),
					Service: projectRoleAssignmentsService,
				},
				&rest.RoleAssignmentsEndpoints{
					AuthFilter: authFilter,
					RoleAssignmentSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/role-assignment.json",
					),
					Service: roleAssignmentsService,
				},
				&rest.SecretsEndpoints{
					AuthFilter: authFilter,
					SecretSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/secret.json",
					),
					Service: secretsService,
				},
				&rest.ServiceAccountEndpoints{
					AuthFilter: authFilter,
					ServiceAccountSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/service-account.json",
					),
					Service: serviceAccountsService,
				},
				&rest.SessionsEndpoints{
					AuthFilter: authFilter,
					Service:    sessionsService,
				},
				&rest.SubstrateEndpoints{
					AuthFilter: authFilter,
					Service:    substrateService,
				},
				&rest.UsersEndpoints{
					AuthFilter: authFilter,
					Service:    usersService,
				},
				&rest.WorkersEndpoints{
					AuthFilter: authFilter,
					WorkerStatusSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/worker-status.json",
					),
					Service: workersService,
				},
				&rest.SystemEndpoints{
					DatabaseClient: database.Client(),
					WriterFactory:  queueWriterFactory,
				},
				&assets.Endpoints{},
			},
			&apiServerConfig,
		)
	}

	// Run it!
	log.Println(apiServer.ListenAndServe(ctx))
}

func testDatabaseConnection(
	ctx context.Context,
	database *mongo.Database,
) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	return retries.ManageRetries(
		ctx, // The retry loop will exit when this context expires
		"database ping",
		0,              // "Infinite" retries
		10*time.Second, // Max backoff
		func() (bool, error) {
			if err := database.Client().Ping(ctx, nil); err != nil {
				return true, err // Retry
			}
			return false, nil // Success
		},
	)
}
