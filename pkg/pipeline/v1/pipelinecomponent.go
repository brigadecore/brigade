package v1

import (
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PipelineComponent describe a pipeline component
type PipelineComponent struct {
	meta_v1.TypeMeta   `json:",inline" yaml:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec               PipelineComponentSpec `json:"spec" yaml:"spec"`
}

//PipelineComponentSpec is the spec for a pipeline component
type PipelineComponentSpec struct {
	Params   []ParameterDefinition     `json:"params"`
	Template PipelineComponentTemplate `json:"template"`
}

//PipelineComponentTemplate defines what is needed to generate the javascript fragment for the brigade script
type PipelineComponentTemplate struct {
	Image   string                    `json:"image"`
	Args    []string                  `json:"args,omitempty"`
	Env     []core_v1.EnvVar          `json:"env,omitempty"`
	Secrets []core_v1.SecretReference `json:"secrets,omitempty"`
	Script  string                    `json:"script,omitempty"`
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
