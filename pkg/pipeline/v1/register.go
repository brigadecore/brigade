package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/Azure/brigade/pkg/pipeline"
)

//SchemeGroupVersion provides the group version
var SchemeGroupVersion = schema.GroupVersion{
	Group:   pipeline.GroupName,
	Version: "v1",
}
var (
	//SchemeBuilder builds a scheme
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	//AddToScheme adds to scheme
	AddToScheme = localSchemeBuilder.AddToScheme
)

func init() {
	localSchemeBuilder.Register(addKnownTypes)
}

//Resource does things to resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds our types to the API scheme by registering
// RadixApplication, RadixApplicationList, RadixDeployment and RadixDeploymentList
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&PipelineComponent{},
		&PipelineComponentList{},
		&PipelineDefinition{},
		&PipelineDefinitionList{},
	)

	// register the type in the scheme
	meta_v1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
