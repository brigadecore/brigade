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
func (c *PipelineClient) GetPipelineDefinitions() ([]v1.PipelineDefinition, error) {
	definitions, err := c.client.PipelineV1().PipelineDefinitions("default").List(meta_v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error occured: %v", err)
		return nil, err
	}
	return definitions.Items, nil
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
