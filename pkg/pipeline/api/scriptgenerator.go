package api

import (
	"fmt"
	"strings"

	v1 "github.com/Azure/brigade/pkg/pipeline/v1"
)

//BrigadeScriptBuilder provides functionality to generate a brigade script based on a pipeline
type BrigadeScriptBuilder struct {
	PipelineConfig *v1.Pipeline
	PipelineAPI    *PipelineClient
	script         strings.Builder
}

func (builder *BrigadeScriptBuilder) generateImports() {
	builder.script.WriteString("const { events, Job, Group } = require('brigadier');")
}

func (builder *BrigadeScriptBuilder) createEvent(triggerEvent v1.PipelineTriggerEvent) *BrigadeEvent {
	event := BrigadeEvent{
		PipelineAPI: builder.PipelineAPI,
		Type:        triggerEvent.Name,
		EventSpec:   triggerEvent.Spec,
	}
	return &event
}

//Script returns the complete javascript
func (builder *BrigadeScriptBuilder) Script() string {
	script := builder.script.String()
	return script
}

//GenerateScript generates javascript based on declarative pipeline configuration
func (c *PipelineClient) GenerateScript(pipelineName string, namespace string) (string, error) {
	fmt.Printf("Generating script from %s in namespace %s", pipelineName, namespace)
	pipeline, err := c.GetPipeline(pipelineName, namespace)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve %s: %v", pipelineName, err)
	}

	builder := &BrigadeScriptBuilder{
		PipelineConfig: pipeline,
		PipelineAPI:    c,
	}

	builder.generateImports()
	for _, triggers := range pipeline.Spec.Events {
		event := builder.createEvent(triggers)
		builder.script.WriteString(event.String())
	}
	// ns := pipeline.Spec.Namespace
	// if ns == "" {
	// 	ns = namespace
	// }

	// definition, err := c.GetPipelineDefinition(pipeline.Spec.Name, ns)
	// if err != nil {
	// 	return nil, fmt.Errorf("Failed to retrieve definition for %s in namespace %s: %v", pipeline.Spec.Name, ns, err)
	// }

	// var scriptBuilder strings.Builder

	// for _, componentSource := range definition.Spec.Pipeline {
	// 	ns = componentSource.Namespace
	// 	if ns == "" {
	// 		ns = namespace
	// 	}

	// 	component, err := c.GetPipelineComponent(componentSource.Name, ns)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Failed to retrieve pipeline component %s: %v", componentSource.Name, err)
	// 	}

	// 	fmt.Fprint(&scriptBuilder, component.Spec.Template)
	// }
	return builder.Script(), nil
}
