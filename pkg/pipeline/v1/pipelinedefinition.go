package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineDefinition describes a pipeline
type PipelineDefinition struct {
	meta_v1.TypeMeta   `json:",inline" yaml:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec               PipelineDefinitionSpec `json:"spec" yaml:"spec"`
}

//PipelineDefinitionSpec is the spec for a pipeline
type PipelineDefinitionSpec struct {
	Params   []ParameterDefinition     `json:"params"`
	Pipeline []PipelineComponentSource `json:"pipeline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//PipelineDefinitionList is a list of pipelines
type PipelineDefinitionList struct {
	meta_v1.TypeMeta `json:",inline" yaml:",inline"`
	meta_v1.ListMeta `json:"metadata" yaml:"metadata"`
	Items            []PipelineDefinition `json:"items" yaml:"items"`
}

//PipelineComponentSource represents a source for the value of a pipeline component
type PipelineComponentSource struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Mandatory bool   `json:"mandatory"`
	//+optional
	ValueFrom *PipelineSource `json:"valueFrom,omitempty"`
}

//PipelineSource represents a source for the value of the pipeline
type PipelineSource struct {
	Param       string            `json:"param,omitempty"`
	PipelineRef PipelineSourceRef `json:"pipelineRef,omitempty"`
}

//PipelineSourceRef represents a reference to a pipeline
type PipelineSourceRef struct {
	Name      string `json:"name,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}
