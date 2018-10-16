package api

import (
	"fmt"
	"strings"
)

//GenerateScript generates javascript based on declarative pipeline configuration
func (c *PipelineClient) GenerateScript(pipelineName string, namespace string) (*string, error) {
	pipeline, err := c.GetPipeline(pipelineName, namespace)
	ns := pipeline.Spec.Namespace
	if ns == "" {
		ns = namespace
	}

	definition, err := c.GetPipelineDefinition(pipeline.Spec.Name, ns)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve definition for %s in namespace %s: %v", pipeline.Spec.Name, ns, err)
	}

	var scriptBuilder strings.Builder

	for _, componentSource := range definition.Spec.Pipeline {
		ns = componentSource.Namespace
		if ns == "" {
			ns = namespace
		}

		component, err := c.GetPipelineComponent(componentSource.Name, ns)
		if err != nil {
			return nil, fmt.Errorf("Failed to retrieve pipeline component %s: %v", componentSource.Name, err)
		}

		fmt.Fprint(&scriptBuilder, component.Spec.Template)
	}
	script := scriptBuilder.String()
	return &script, nil
}
