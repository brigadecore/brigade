package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Pipeline describes a pipeline
type Pipeline struct {
	meta_v1.TypeMeta   `json:",inline" yaml:",inline"`
	meta_v1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec               PipelineSpec `json:"spec" yaml:"spec"`
}

//PipelineSpec describes the pipelines for various triggers
type PipelineSpec struct {
	Events []PipelineTriggerEvent `json:"events"`
}

//PipelineTriggerEvent describes which pipeline to run on a trigger ie. push, pull_request, etc
type PipelineTriggerEvent struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Spec        PipelineEventSpec `json:"spec"`
}

//PipelineEventSpec describes a pipeline for a trigger event
type PipelineEventSpec struct {
	Name      string              `json:"name"`
	Namespace string              `json:"namespace,omitempty"`
	Params    []PipelineParameter `json:"params"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//PipelineList is a list of pipeline components
type PipelineList struct {
	meta_v1.TypeMeta `json:",inline" yaml:",inline"`
	meta_v1.ListMeta `json:"metadata" yaml:"metadata"`
	Items            []Pipeline `json:"items" yaml:"items"`
}

//PipelineParameter describes a input value to the specified pipeline
type PipelineParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
