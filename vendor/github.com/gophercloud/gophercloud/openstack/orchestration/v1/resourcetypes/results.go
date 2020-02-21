package resourcetypes

import (
	"github.com/gophercloud/gophercloud"
)

// ResourceTypeSummary contains the result of listing an available resource
// type.
type ResourceTypeSummary struct {
	ResourceType string `json:"resource_type"`
	Description  string `json:"description"`
}

// PropertyType represents the expected type of a property or attribute value.
type PropertyType string

const (
	// StringProperty indicates a string property type.
	StringProperty PropertyType = "string"
	// IntegerProperty indicates an integer property type.
	IntegerProperty PropertyType = "integer"
	// NumberProperty indicates a number property type. It may be an integer or
	// float.
	NumberProperty PropertyType = "number"
	// BooleanProperty indicates a boolean property type.
	BooleanProperty PropertyType = "boolean"
	// MapProperty indicates a map property type.
	MapProperty PropertyType = "map"
	// ListProperty indicates a list property type.
	ListProperty PropertyType = "list"
	// UntypedProperty indicates a property that could have any type.
	UntypedProperty PropertyType = "any"
)

// AttributeSchema is the schema of a resource attribute
type AttributeSchema struct {
	Description string       `json:"description,omitempty"`
	Type        PropertyType `json:"type"`
}

// MinMaxConstraint is a type of constraint with minimum and maximum values.
// This is used for both Range and Length constraints.
type MinMaxConstraint struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

// ModuloConstraint constrains an integer to have a certain value given a
// particular modulus.
type ModuloConstraint struct {
	Step   int `json:"step,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ConstraintSchema describes all possible types of constraints. Besides the
// description, only one other field is ever set at a time.
type ConstraintSchema struct {
	Description      string            `json:"description,omitempty"`
	Range            *MinMaxConstraint `json:"range,omitempty"`
	Length           *MinMaxConstraint `json:"length,omitempty"`
	Modulo           *ModuloConstraint `json:"modulo,omitempty"`
	AllowedValues    *[]interface{}    `json:"allowed_values,omitempty"`
	AllowedPattern   *string           `json:"allowed_pattern,omitempty"`
	CustomConstraint *string           `json:"custom_constraint,omitempty"`
}

// PropertySchema is the schema of a resource property.
type PropertySchema struct {
	Type          PropertyType              `json:"type"`
	Description   string                    `json:"description,omitempty"`
	Default       interface{}               `json:"default,omitempty"`
	Constraints   []ConstraintSchema        `json:"constraints,omitempty"`
	Required      bool                      `json:"required"`
	Immutable     bool                      `json:"immutable"`
	UpdateAllowed bool                      `json:"update_allowed"`
	Schema        map[string]PropertySchema `json:"schema,omitempty"`
}

// SupportStatusDetails contains information about the support status of the
// resource and its history.
type SupportStatusDetails struct {
	Status         SupportStatus         `json:"status"`
	Message        string                `json:"message,omitempty"`
	Version        string                `json:"version,omitempty"`
	PreviousStatus *SupportStatusDetails `json:"previous_status,omitempty"`
}

// ResourceSchema is the schema for a resource type, its attributes, and
// properties.
type ResourceSchema struct {
	ResourceType  string                     `json:"resource_type"`
	SupportStatus SupportStatusDetails       `json:"support_status"`
	Attributes    map[string]AttributeSchema `json:"attributes"`
	Properties    map[string]PropertySchema  `json:"properties"`
}

// ListResult represents the result of a List operation.
type ListResult struct {
	gophercloud.Result
}

// Extract returns a slice of ResourceTypeSummary objects and is called after
// a List operation.
func (r ListResult) Extract() (rts []ResourceTypeSummary, err error) {
	var full struct {
		ResourceTypes []ResourceTypeSummary `json:"resource_types"`
	}
	err = r.ExtractInto(&full)
	if err == nil {
		rts = full.ResourceTypes
		return
	}

	var basic struct {
		ResourceTypes []string `json:"resource_types"`
	}
	err2 := r.ExtractInto(&basic)
	if err2 == nil {
		err = nil
		rts = make([]ResourceTypeSummary, len(basic.ResourceTypes))
		for i, n := range basic.ResourceTypes {
			rts[i] = ResourceTypeSummary{ResourceType: n}
		}
	}
	return
}

// GetSchemaResult represents the result of a GetSchema operation.
type GetSchemaResult struct {
	gophercloud.Result
}

// Extract returns a ResourceSchema object and is called after a GetSchema
// operation.
func (r GetSchemaResult) Extract() (rts ResourceSchema, err error) {
	err = r.ExtractInto(&rts)
	return
}

// TemplateResult represents the result of a Template get operation.
type TemplateResult struct {
	gophercloud.Result
}

// Extract returns a Template object and is called after a Template get
// operation.
func (r TemplateResult) Extract() (template map[string]interface{}, err error) {
	err = r.ExtractInto(&template)
	return
}
