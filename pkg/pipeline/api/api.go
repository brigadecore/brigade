package api

import (
	"fmt"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipeclient "github.com/Azure/brigade/pkg/pipeline/client/clientset/versioned"
	"github.com/Azure/brigade/pkg/pipeline/v1"
	"k8s.io/client-go/rest"
)

//PipelineClient provides functionality for working with pipelines
type PipelineClient struct {
	client pipeclient.Interface
}

//GetPipelineDefinitions returns a list of all registered pipeline definitions
func (c *PipelineClient) GetPipelineDefinitions(namespace string) ([]v1.PipelineDefinition, error) {
	definitions, err := c.client.PipelineV1().PipelineDefinitions(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error occured: %v", err)
		return nil, err
	}
	return definitions.Items, nil
}

//GetPipelineDefinition retrieves the specified pipeline definition
func (c *PipelineClient) GetPipelineDefinition(name string, namespace string) (*v1.PipelineDefinition, error) {
	definition, err := c.client.PipelineV1().PipelineDefinitions(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		fmt.Printf("Error occured: %v", err)
		return nil, err
	}
	return definition, nil
}

//GetPipelineComponents returns a list of all registered pipeline components
func (c *PipelineClient) GetPipelineComponents(namespace string) ([]v1.PipelineComponent, error) {
	components, err := c.client.PipelineV1().PipelineComponents(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error occured: %v", err)
		return nil, err
	}

	return components.Items, nil
}

//GetPipelineComponent retrieves the specified pipeline component
func (c *PipelineClient) GetPipelineComponent(name string, namespace string) (*v1.PipelineComponent, error) {
	component, err := c.client.PipelineV1().PipelineComponents(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return component, nil
}

//GetPipelines retrieves all registered pipelines
func (c *PipelineClient) GetPipelines(namespace string) ([]v1.Pipeline, error) {
	pipelines, err := c.client.PipelineV1().Pipelines(namespace).List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pipelines.Items, nil
}

//GetPipeline retrieve specified pipeline
func (c *PipelineClient) GetPipeline(name string, namespace string) (*v1.Pipeline, error) {
	pipeline, err := c.client.PipelineV1().Pipelines(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pipeline, nil
}

//GetPipelineComponentFrom retrieves a PipelineComponent or other Pipeline from the specified pipeline definition
func (c *PipelineClient) GetPipelineComponentFrom(definition *v1.PipelineDefinition, name string) (meta_v1.Object, error) {
	for _, componentSource := range definition.Spec.Phases {
		if componentSource.Name == name && componentSource.ValueFrom == nil {
			return c.GetPipelineComponent(name, componentSource.Namespace)
		}
	}
	return nil, fmt.Errorf("Could not find component %s", name)
}

//New creates a new PipelineClient
func New(config *rest.Config) (*PipelineClient, error) {
	clientset, err := pipeclient.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create pipeline client: %v", err)
		return nil, err
	}

	client := &PipelineClient{
		client: clientset,
	}

	return client, nil
}
