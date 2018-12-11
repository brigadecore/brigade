package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Azure/brigade/pkg/api"
	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage/kube"

	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/go-openapi/spec"

	"k8s.io/api/core/v1"
)

var (
	apiPort    string
	kubeconfig string
	master     string
	namespace  string
	verbose    bool
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	flag.StringVar(&master, "master", "", "master url")
	flag.StringVar(&namespace, "namespace", defaultNamespace(), "kubernetes namespace")
	flag.StringVar(&apiPort, "api-port", defaultAPIPort(), "TCP port to use for brigade-api")
	flag.BoolVar(&verbose, "verbose", false, "enables detailed logging of http request matching and filter invocation")
}

type jobService struct {
	server api.API
}

type buildService struct {
	server api.API
}

type projectService struct {
	server api.API
}

type healthService struct {
}

func (js jobService) WebService() *restful.WebService {
	ws := new(restful.WebService)
	j := js.server.Job()

	ws.
		Path("/v1/job").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML, "plain/text", "text/javascript")

	tags := []string{"job"}

	ws.Route(ws.GET("/{id}").To(j.Get).
		Doc("get a job").
		Param(ws.PathParameter("id", "identifier of the job").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(brigade.Job{}). // on the response
		Returns(200, "OK", brigade.Job{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/{id}/logs").To(j.Logs).
		Doc("get job logs").
		Param(ws.PathParameter("id", "identifier of the job").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]byte{}). // on the response
		Returns(200, "OK", []byte{}).
		Returns(404, "Not Found", nil))

	return ws
}

func (bs buildService) WebService() *restful.WebService {
	ws := new(restful.WebService)
	b := bs.server.Build()

	ws.
		Path("/v1/build").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML, "plain/text", "text/javascript")

	tags := []string{"build"}

	ws.Route(ws.GET("/{id}").To(b.Get).
		Doc("get a build").
		Param(ws.PathParameter("id", "id of the build").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(brigade.Build{}).
		Returns(200, "OK", brigade.Build{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/{id}/jobs").To(b.Jobs).
		Doc("get jobs of a build").
		Param(ws.PathParameter("id", "id of the build").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]brigade.Job{}).
		Returns(200, "OK", []brigade.Job{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/{id}/logs").To(b.Logs).
		Doc("get logs of a build").
		Param(ws.PathParameter("id", "id of the build").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]byte{}).
		Returns(200, "OK", []byte{}).
		Returns(404, "Not Found", nil))

	return ws
}

func (ps projectService) WebService() *restful.WebService {
	ws := new(restful.WebService)
	p := ps.server.Project()

	ws.
		Path("/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML, "plain/text", "text/javascript")

	tags := []string{"projects"}

	ws.Route(ws.GET("/projects").To(p.List).
		Doc("get all projects").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]brigade.Project{}).
		Returns(200, "OK", []brigade.Project{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/project/{id}").To(p.Get).
		Param(ws.PathParameter("id", "id of the project").DataType("string")).
		Doc("get a project").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes(brigade.Project{}).
		Returns(200, "OK", brigade.Project{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/project/{id}/builds").To(p.Builds).
		Doc("get list of builds for a project").
		Param(ws.PathParameter("id", "id of the project").DataType("string")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]brigade.Build{}).
		Returns(200, "OK", []brigade.Build{}).
		Returns(404, "Not Found", nil))

	ws.Route(ws.GET("/projects-build").To(p.ListWithLatestBuild).
		Doc("lists the projects with the latest builds attached.").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]api.ProjectBuildSummary{}).
		Returns(200, "OK", []api.ProjectBuildSummary{}).
		Returns(404, "Not Found", nil))

	return ws
}

func (hs healthService) WebService() *restful.WebService {
	ws := new(restful.WebService)

	ws.
		Path("/healthz").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"healthz"}

	ws.Route(ws.GET("/").To(api.Healthz).
		Doc("get health status").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]byte{}).
		Returns(200, "OK", []byte{}).
		Returns(404, "Not Found", nil))

	return ws
}

func main() {
	flag.Parse()

	restful.EnableTracing(verbose)

	clientset, err := kube.GetClient(master, kubeconfig)
	if err != nil {
		log.Fatalf("error creating kubernetes client (%s)", err)
		return
	}

	storage := kube.New(clientset, namespace)
	storageServer := api.New(storage)

	j := jobService{server: storageServer}
	b := buildService{server: storageServer}
	p := projectService{server: storageServer}
	h := healthService{}

	restful.DefaultContainer.Add(j.WebService())
	restful.DefaultContainer.Add(b.WebService())
	restful.DefaultContainer.Add(p.WebService())
	restful.DefaultContainer.Add(h.WebService())
	restful.DefaultContainer.Filter(NCSACommonLogFormatLogger())

	config := restfulspec.Config{
		WebServices:                   restful.RegisteredWebServices(),
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	restful.DefaultContainer.Add(restfulspec.NewOpenAPIService(config))

	cors := restful.CrossOriginResourceSharing{
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		CookiesAllowed: false,
		Container:      restful.DefaultContainer}
	restful.DefaultContainer.Filter(cors.Filter)

	formattedAPIPort := fmt.Sprintf(":%v", apiPort)

	log.Printf("Get the API using %s/apidocs.json", formattedAPIPort)
	hserver := &http.Server{Addr: formattedAPIPort, Handler: restful.DefaultContainer}
	log.Fatal(hserver.ListenAndServe())
}

func defaultNamespace() string {
	if ns, ok := os.LookupEnv("BRIGADE_NAMESPACE"); ok {
		return ns
	}
	return v1.NamespaceDefault
}

func defaultAPIPort() string {
	if port, ok := os.LookupEnv("BRIGADE_API_PORT"); ok {
		return port
	}
	return "7745"
}

func enrichSwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "Brigade API",
			Description: "Resources for Jobs, Projects, Builds",
			License: &spec.License{
				Name: "MIT",
				URL:  "http://mit.org",
			},
			Version: "1.0.0",
		},
	}
}
