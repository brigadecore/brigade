package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineComponent describe an application
type PipelineComponent struct {
	meta_v1.TypeMeta   `json:",inline" yaml:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec               PipelineComponentSpec `json:"spec" yaml:"spec"`
}

//PipelineComponentSpec is the spec for an application
type PipelineComponentSpec struct {
	Params []ParameterDefinition `json:"params"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//PipelineComponentList is a list of pipeline components
type PipelineComponentList struct {
	meta_v1.TypeMeta `json:",inline" yaml:",inline"`
	meta_v1.ListMeta `json:"metadata" yaml:"metadata"`
	Items            []PipelineComponent `json:"items" yaml:"items"`
}

//ParameterDefinition defines an configurable parameter for the pipeline component
type ParameterDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
