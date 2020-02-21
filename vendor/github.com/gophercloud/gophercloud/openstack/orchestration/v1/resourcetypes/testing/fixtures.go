package testing

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/orchestration/v1/resourcetypes"
	th "github.com/gophercloud/gophercloud/testhelper"
	fake "github.com/gophercloud/gophercloud/testhelper/client"
)

const BasicListOutput = `
{
    "resource_types": [
		"OS::Nova::Server",
		"OS::Heat::Stack"
    ]
}
`

var BasicListExpected = []resourcetypes.ResourceTypeSummary{
	resourcetypes.ResourceTypeSummary{
		ResourceType: "OS::Nova::Server",
	},
	resourcetypes.ResourceTypeSummary{
		ResourceType: "OS::Heat::Stack",
	},
}

const FullListOutput = `
{
    "resource_types": [
        {
            "description": "A Nova Server",
			"resource_type": "OS::Nova::Server"
        },
        {
            "description": "A Heat Stack",
			"resource_type": "OS::Heat::Stack"
        }
    ]
}
`

var FullListExpected = []resourcetypes.ResourceTypeSummary{
	resourcetypes.ResourceTypeSummary{
		ResourceType: "OS::Nova::Server",
		Description:  "A Nova Server",
	},
	resourcetypes.ResourceTypeSummary{
		ResourceType: "OS::Heat::Stack",
		Description:  "A Heat Stack",
	},
}

const listFilterRegex = "OS::Heat::.*"
const FilteredListOutput = `
{
    "resource_types": [
        {
            "description": "A Heat Stack",
			"resource_type": "OS::Heat::Stack"
        }
    ]
}
`

var FilteredListExpected = []resourcetypes.ResourceTypeSummary{
	resourcetypes.ResourceTypeSummary{
		ResourceType: "OS::Heat::Stack",
		Description:  "A Heat Stack",
	},
}

// HandleListSuccessfully creates an HTTP handler at `/resource_types`
// on the test handler mux that responds with a `List` response.
func HandleListSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/resource_types",
		func(w http.ResponseWriter, r *http.Request) {
			th.TestMethod(t, r, "GET")
			th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
			th.TestHeader(t, r, "Accept", "application/json")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			r.ParseForm()
			var output string
			if r.Form.Get("with_description") == "true" {
				if r.Form.Get("name") == listFilterRegex {
					output = FilteredListOutput
				} else {
					output = FullListOutput
				}
			} else {
				output = BasicListOutput
			}
			fmt.Fprint(w, output)
		})
}

var glanceImageConstraint = "glance.image"

var GetSchemaExpected = resourcetypes.ResourceSchema{
	ResourceType: "OS::Test::TestServer",
	SupportStatus: resourcetypes.SupportStatusDetails{
		Status:  resourcetypes.SupportStatusDeprecated,
		Message: "Bye bye.",
		Version: "10.0.0",
		PreviousStatus: &resourcetypes.SupportStatusDetails{
			Status: resourcetypes.SupportStatusSupported,
		},
	},
	Attributes: map[string]resourcetypes.AttributeSchema{
		"show": resourcetypes.AttributeSchema{
			Description: "Detailed information about resource.",
			Type:        resourcetypes.MapProperty,
		},
		"tags": resourcetypes.AttributeSchema{
			Description: "Tags from the server.",
			Type:        resourcetypes.ListProperty,
		},
		"name": resourcetypes.AttributeSchema{
			Description: "Name of the server.",
			Type:        resourcetypes.StringProperty,
		},
	},
	Properties: map[string]resourcetypes.PropertySchema{
		"name": resourcetypes.PropertySchema{
			Type:          resourcetypes.StringProperty,
			Description:   "Server name.",
			UpdateAllowed: true,
		},
		"image": resourcetypes.PropertySchema{
			Type:        resourcetypes.StringProperty,
			Description: "The ID or name of the image to boot with.",
			Required:    true,
			Constraints: []resourcetypes.ConstraintSchema{
				resourcetypes.ConstraintSchema{
					CustomConstraint: &glanceImageConstraint,
				},
			},
		},
		"block_device_mapping": resourcetypes.PropertySchema{
			Type:        resourcetypes.ListProperty,
			Description: "Block device mappings for this server.",
			Schema: map[string]resourcetypes.PropertySchema{
				"*": resourcetypes.PropertySchema{
					Type: resourcetypes.MapProperty,
					Schema: map[string]resourcetypes.PropertySchema{
						"ephemeral_format": resourcetypes.PropertySchema{
							Type:        resourcetypes.StringProperty,
							Description: "The format of the local ephemeral block device.",
							Constraints: []resourcetypes.ConstraintSchema{
								resourcetypes.ConstraintSchema{
									AllowedValues: &[]interface{}{
										"ext3", "ext4", "xfs",
									},
								},
							},
						},
						"ephemeral_size": resourcetypes.PropertySchema{
							Type:        resourcetypes.IntegerProperty,
							Description: "The size of the local ephemeral block device, in GB.",
							Constraints: []resourcetypes.ConstraintSchema{
								resourcetypes.ConstraintSchema{
									Range: &resourcetypes.MinMaxConstraint{
										Min: 1,
									},
								},
							},
						},
						"delete_on_termination": resourcetypes.PropertySchema{
							Type:        resourcetypes.BooleanProperty,
							Description: "Delete volume on server termination.",
							Default:     true,
							Immutable:   true,
						},
					},
				},
			},
		},
		"image_update_policy": resourcetypes.PropertySchema{
			Type:        resourcetypes.StringProperty,
			Description: "Policy on how to apply an image-id update.",
			Default:     "REBUILD",
			Constraints: []resourcetypes.ConstraintSchema{
				resourcetypes.ConstraintSchema{
					AllowedValues: &[]interface{}{
						"REBUILD", "REPLACE",
					},
				},
			},
			UpdateAllowed: true,
		},
	},
}

const GetSchemaOutput = `
{
  "resource_type": "OS::Test::TestServer",
  "support_status": {
    "status": "DEPRECATED",
    "message": "Bye bye.",
    "version": "10.0.0",
    "previous_status": {
      "status": "SUPPORTED",
      "message": null,
      "version": null,
      "previous_status": null
    }
  },
  "attributes": {
    "show": {
      "type": "map",
      "description": "Detailed information about resource."
    },
    "tags": {
      "type": "list",
      "description": "Tags from the server."
    },
    "name": {
      "type": "string",
      "description": "Name of the server."
    }
  },
  "properties": {
    "name": {
      "update_allowed": true,
      "required": false,
      "type": "string",
      "description": "Server name.",
      "immutable": false
    },
    "image": {
      "description": "The ID or name of the image to boot with.",
      "required": true,
      "update_allowed": false,
      "type": "string",
      "immutable": false,
      "constraints": [
        {
          "custom_constraint": "glance.image"
        }
      ]
    },
    "block_device_mapping": {
      "description": "Block device mappings for this server.",
      "required": false,
      "update_allowed": false,
      "type": "list",
      "immutable": false,
      "schema": {
        "*": {
          "update_allowed": false,
          "required": false,
          "type": "map",
          "immutable": false,
          "schema": {
            "ephemeral_format": {
              "description": "The format of the local ephemeral block device.",
              "required": false,
              "update_allowed": false,
              "type": "string",
              "immutable": false,
              "constraints": [
                {
                  "allowed_values": [
                    "ext3",
                    "ext4",
                    "xfs"
                  ]
                }
              ]
            },
            "ephemeral_size": {
              "description": "The size of the local ephemeral block device, in GB.",
              "required": false,
              "update_allowed": false,
              "type": "integer",
              "immutable": false,
              "constraints": [
                {
                  "range": {
                    "min": 1
                  }
                }
              ]
            },
            "delete_on_termination": {
              "update_allowed": false,
              "default": true,
              "required": false,
              "type": "boolean",
              "description": "Delete volume on server termination.",
              "immutable": true
            }
          }
        }
      }
    },
    "image_update_policy": {
      "description": "Policy on how to apply an image-id update.",
      "default": "REBUILD",
      "required": false,
      "update_allowed": true,
      "type": "string",
      "immutable": false,
      "constraints": [
        {
          "allowed_values": [
            "REBUILD",
            "REPLACE"
          ]
        }
      ]
    }
  }
}
`

// HandleGetSchemaSuccessfully creates an HTTP handler at
// `/resource_types/OS::Test::TestServer` on the test handler mux that
// responds with a `GetSchema` response.
func HandleGetSchemaSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/resource_types/OS::Test::TestServer",
		func(w http.ResponseWriter, r *http.Request) {
			th.TestMethod(t, r, "GET")
			th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
			th.TestHeader(t, r, "Accept", "application/json")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, GetSchemaOutput)
		})
}

const GenerateTemplateOutput = `
{
  "outputs": {
    "OS::stack_id": {
      "value": {
        "get_resource": "NoneResource"
      }
    },
    "show": {
      "description": "Detailed information about resource.",
      "value": {
        "get_attr": [
          "NoneResource",
          "show"
        ]
      }
    }
  },
  "heat_template_version": "2016-10-14",
  "description": "Initial template of NoneResource",
  "parameters": {},
  "resources": {
    "NoneResource": {
      "type": "OS::Heat::None",
      "properties": {}
    }
  }
}
`

// HandleGenerateTemplateSuccessfully creates an HTTP handler at
// `/resource_types/OS::Heat::None/template` on the test handler mux that
// responds with a template.
func HandleGenerateTemplateSuccessfully(t *testing.T) {
	th.Mux.HandleFunc("/resource_types/OS::Heat::None/template",
		func(w http.ResponseWriter, r *http.Request) {
			th.TestMethod(t, r, "GET")
			th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
			th.TestHeader(t, r, "Accept", "application/json")

			w.Header().Set("Content-Type", "application/json")
			r.ParseForm()
			if r.Form.Get("template_type") == "hot" {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, GenerateTemplateOutput)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		})
}
