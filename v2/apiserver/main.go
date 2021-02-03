package main

// nolint: lll
import (
	"log"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authn"
	authnMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/authn/mongodb"
	authnREST "github.com/brigadecore/brigade/v2/apiserver/internal/authn/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	authzMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/authz/mongodb"
	authzREST "github.com/brigadecore/brigade/v2/apiserver/internal/authz/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	coreKubernetes "github.com/brigadecore/brigade/v2/apiserver/internal/core/kubernetes"
	coreMongodb "github.com/brigadecore/brigade/v2/apiserver/internal/core/mongodb"
	coreREST "github.com/brigadecore/brigade/v2/apiserver/internal/core/rest"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue/amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	sysAuthn "github.com/brigadecore/brigade/v2/apiserver/internal/system/authn"
	sysAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/system/authz"
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
	var coolLogsStore core.LogsStore
	var eventsStore core.EventsStore
	var jobsStore core.JobsStore
	var projectsStore core.ProjectsStore
	var roleAssignmentsStore authz.RoleAssignmentsStore
	var secretsStore core.SecretsStore
	var serviceAccountsStore authn.ServiceAccountsStore
	var sessionsStore authn.SessionsStore
	var usersStore authn.UsersStore
	var warmLogsStore core.LogsStore
	var workersStore core.WorkersStore
	{
		database, err := database(ctx)
		if err != nil {
			log.Fatal(err)
		}
		coolLogsStore = coreMongodb.NewLogsStore(database)
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
		roleAssignmentsStore = authzMongodb.NewRoleAssignmentsStore(database)
		secretsStore = coreKubernetes.NewSecretsStore(kubeClient)
		serviceAccountsStore, err = authnMongodb.NewServiceAccountsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		sessionsStore, err = authnMongodb.NewSessionsStore(database)
		if err != nil {
			log.Fatal(err)
		}
		usersStore, err = authnMongodb.NewUsersStore(database)
		if err != nil {
			log.Fatal(err)
		}
		warmLogsStore = coreKubernetes.NewLogsStore(kubeClient)
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

	// Authorizer
	authorizer := sysAuthz.NewAuthorizer(roleAssignmentsStore)

	// Events service
	eventsService := core.NewEventsService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)

	// Jobs service
	jobsService := core.NewJobsService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)

	// Logs service
	logsService := core.NewLogsService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		warmLogsStore,
		coolLogsStore,
	)

	// Projects service
	projectsService := core.NewProjectsService(
		authorizer.Authorize,
		projectsStore,
		eventsStore,
		roleAssignmentsStore,
		substrate,
	)

	// ProjectRoleAssignments service
	projectRoleAssignmentsService := core.NewProjectRoleAssignmentsService(
		authorizer.Authorize,
		projectsStore,
		usersStore,
		serviceAccountsStore,
		roleAssignmentsStore,
	)

	// Roles service
	roleAssignmentsService := authz.NewRoleAssignmentsService(
		authorizer.Authorize,
		usersStore,
		serviceAccountsStore,
		roleAssignmentsStore,
	)

	// ServiceAccounts service
	serviceAccountsService :=
		authn.NewServiceAccountsService(authorizer.Authorize, serviceAccountsStore)

	// Secrets service
	secretsService := core.NewSecretsService(
		authorizer.Authorize,
		projectsStore,
		secretsStore,
	)

	// Session service
	var sessionsService authn.SessionsService
	{
		config, err := sessionsServiceConfig(ctx)
		if err != nil {
			log.Fatal(err)
		}
		sessionsService = authn.NewSessionsService(
			sessionsStore,
			usersStore,
			&config,
		)
	}

	// Substrate service
	substrateService := core.NewSubstrateService(authorizer.Authorize, substrate)

	// Users service
	usersService := authn.NewUsersService(
		authorizer.Authorize,
		usersStore,
		sessionsStore,
	)

	// Workers service
	workersService := core.NewWorkersService(
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
		authFilter := sysAuthn.NewTokenAuthFilter(
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
				&coreREST.EventsEndpoints{
					AuthFilter: authFilter,
					EventSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/event.json",
					),
					Service: eventsService,
				},
				&coreREST.JobsEndpoints{
					AuthFilter: authFilter,
					JobSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/job.json",
					),
					JobStatusSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/job-status.json",
					),
					Service: jobsService,
				},
				&coreREST.LogsEndpoints{
					AuthFilter: authFilter,
					Service:    logsService,
				},
				&coreREST.ProjectsEndpoints{
					AuthFilter: authFilter,
					ProjectSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/project.json",
					),
					Service: projectsService,
				},
				&coreREST.ProjectRoleAssignmentsEndpoints{
					AuthFilter: authFilter,
					ProjectRoleAssignmentSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/project-role-assignment.json",
					),
					Service: projectRoleAssignmentsService,
				},
				&authzREST.RoleAssignmentsEndpoints{
					AuthFilter: authFilter,
					RoleAssignmentSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/role-assignment.json",
					),
					Service: roleAssignmentsService,
				},
				&coreREST.SecretsEndpoints{
					AuthFilter: authFilter,
					SecretSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/secret.json",
					),
					Service: secretsService,
				},
				&authnREST.ServiceAccountEndpoints{
					AuthFilter: authFilter,
					ServiceAccountSchemaLoader: gojsonschema.NewReferenceLoader(
						"file:///brigade/schemas/service-account.json",
					),
					Service: serviceAccountsService,
				},
				&authnREST.SessionsEndpoints{
					AuthFilter: authFilter,
					Service:    sessionsService,
				},
				&coreREST.SubstrateEndpoints{
					AuthFilter: authFilter,
					Service:    substrateService,
				},
				&authnREST.UsersEndpoints{
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
