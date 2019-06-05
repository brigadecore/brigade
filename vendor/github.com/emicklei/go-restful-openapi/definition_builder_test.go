package restfulspec

import (
	"testing"

	"github.com/go-openapi/spec"
)

type Apple struct {
	Species string
	Volume  int `json:"vol"`
	Things  *[]string
}

func TestAppleDef(t *testing.T) {
	db := definitionBuilder{Definitions: spec.Definitions{}, Config: Config{}}
	db.addModelFrom(Apple{})

	if got, want := len(db.Definitions), 1; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema := db.Definitions["restfulspec.Apple"]
	if got, want := len(schema.Required), 3; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Required[0], "Species"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Required[1], "vol"; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.ID, ""; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Properties["Things"].Items.Schema.Type.Contains("string"), true; got != want {
		t.Errorf("got %v want %v", got, want)
	}
	if got, want := schema.Properties["Things"].Items.Schema.Ref.String(), ""; got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

type MyDictionaryResponse struct {
	Dictionary1 map[string]DictionaryValue `json:"dictionary1"`
	Dictionary2 map[string]interface{}     `json:"dictionary2"`
}
type DictionaryValue struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
}

func TestDictionarySupport(t *testing.T) {
	db := definitionBuilder{Definitions: spec.Definitions{}, Config: Config{}}
	db.addModelFrom(MyDictionaryResponse{})

	// Make sure that only the types that we want were created.
	if got, want := len(db.Definitions), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema, schemaFound := db.Definitions["restfulspec.MyDictionaryResponse"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Required[0], "dictionary1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Required[1], "dictionary2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Properties), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Properties["dictionary1"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.AdditionalProperties.Schema.SchemaProps.Ref.String(), "#/definitions/restfulspec.DictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
			if property, found := schema.Properties["dictionary2"]; !found {
				t.Errorf("could not find property")
			} else {
				if property.AdditionalProperties != nil {
					t.Errorf("unexpected additional properties")
				}
			}
		}
	}

	schema, schemaFound = db.Definitions["restfulspec.DictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Required[0], "key1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Required[1], "key2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
	}
}

type MyRecursiveDictionaryResponse struct {
	Dictionary1 map[string]RecursiveDictionaryValue `json:"dictionary1"`
}
type RecursiveDictionaryValue struct {
	Key1 string                              `json:"key1"`
	Key2 map[string]RecursiveDictionaryValue `json:"key2"`
}

func TestRecursiveDictionarySupport(t *testing.T) {
	db := definitionBuilder{Definitions: spec.Definitions{}, Config: Config{}}
	db.addModelFrom(MyRecursiveDictionaryResponse{})

	// Make sure that only the types that we want were created.
	if got, want := len(db.Definitions), 2; got != want {
		t.Errorf("got %v want %v", got, want)
	}

	schema, schemaFound := db.Definitions["restfulspec.MyRecursiveDictionaryResponse"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Required), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Required[0], "dictionary1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Properties), 1; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Properties["dictionary1"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.AdditionalProperties.Schema.SchemaProps.Ref.String(), "#/definitions/restfulspec.RecursiveDictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
		}
	}

	schema, schemaFound = db.Definitions["restfulspec.RecursiveDictionaryValue"]
	if !schemaFound {
		t.Errorf("could not find schema")
	} else {
		if got, want := len(schema.Required), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if got, want := schema.Required[0], "key1"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
			if got, want := schema.Required[1], "key2"; got != want {
				t.Errorf("got %v want %v", got, want)
			}
		}
		if got, want := len(schema.Properties), 2; got != want {
			t.Errorf("got %v want %v", got, want)
		} else {
			if property, found := schema.Properties["key1"]; !found {
				t.Errorf("could not find property")
			} else {
				if property.AdditionalProperties != nil {
					t.Errorf("unexpected additional properties")
				}
			}
			if property, found := schema.Properties["key2"]; !found {
				t.Errorf("could not find property")
			} else {
				if got, want := property.AdditionalProperties.Schema.SchemaProps.Ref.String(), "#/definitions/restfulspec.RecursiveDictionaryValue"; got != want {
					t.Errorf("got %v want %v", got, want)
				}
			}
		}
	}
}
