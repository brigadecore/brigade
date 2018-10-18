package api

import (
	"fmt"
	"strings"

	"github.com/Azure/brigade/pkg/pipeline/v1"
	"github.com/iancoleman/strcase"
)

//BrigadeEvent represents a trigger event
type BrigadeEvent struct {
	PipelineAPI *PipelineClient
	Type        string
	EventSpec   v1.PipelineEventSpec
	builder     strings.Builder
	params      map[string]string
}

func (e *BrigadeEvent) String() string {
	fmt.Fprintf(&e.builder, "\n\nevents.on('%s', (brigadeEvent, project) => {\n", e.Type)

	definition, err := e.PipelineAPI.GetPipelineDefinition(e.EventSpec.Name, e.EventSpec.Namespace)
	if err != nil {
		fmt.Printf("Failed to retrieve pipeline definition %s: %v", e.EventSpec.Name, err)
		return ""
	}
	e.gatherParameters(definition)
	e.createJobs(definition)

	e.builder.WriteString("\n});\n")

	script := e.builder.String()
	script = e.insertParameterValues(script)
	return script
}

func (e *BrigadeEvent) insertParameterValues(script string) string {
	for param, paramValue := range e.params {
		script = strings.Replace(script, fmt.Sprintf("{{%s}}", param), paramValue, -1)
	}

	return script
}

func (e *BrigadeEvent) gatherParameters(definition *v1.PipelineDefinition) {
	e.params = make(map[string]string)
	for _, param := range definition.Spec.Params {
		paramValue := ""
		for _, p := range e.EventSpec.Params {
			if p.Name == param.Name {
				paramValue = p.Value
				break
			}
		}

		e.params[param.Name] = paramValue
	}
}

func (e *BrigadeEvent) createJobs(definition *v1.PipelineDefinition) {
	var jobList []string
	for _, phase := range definition.Spec.Phases {
		jobName, err := e.createJob(definition, phase)
		if err != nil {
			fmt.Printf("Error creating job: %v", err)
			break
		}

		jobList = append(jobList, jobName)
	}
	e.createPipeline(jobList)
}

func (e *BrigadeEvent) createJob(definition *v1.PipelineDefinition, componentSource v1.PipelineComponentSource) (string, error) {
	component, err := e.PipelineAPI.GetPipelineComponentFrom(definition, componentSource.Name)
	if err != nil {

		return "", fmt.Errorf("Error building script: %v", err)
	}

	jobName := ""
	switch t := component.(type) {
	case *v1.PipelineComponent:
		jobName = strcase.ToLowerCamel(t.Name)
		fmt.Fprintf(&e.builder, "\tvar %s = new Job('%s');\n", jobName, t.Name)
		fmt.Fprintf(&e.builder, "\t%s.image = '%s';\n", jobName, t.Spec.Template.Image)
		fmt.Fprintf(&e.builder, "\t%s.args = [\n", jobName)
		for _, arg := range t.Spec.Template.Args {
			fmt.Fprintf(&e.builder, "\t\t'%s',\n", arg)
		}
		e.builder.WriteString("\t];")

	case *v1.PipelineDefinition:
	}

	return jobName, nil
}

func (e *BrigadeEvent) createPipeline(jobs []string) {
	e.builder.WriteString("\n\tvar pipeline = new Group();\n")
	for _, jobName := range jobs {
		fmt.Fprintf(&e.builder, "\tpipeline.add(%s);\n", jobName)
	}
	e.builder.WriteString("\tpipeline.runEach();")
}
