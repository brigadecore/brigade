package main

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"
	spec2 "github.com/go-openapi/spec"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestAppleDef(t *testing.T) {

	raw, err := ioutil.ReadFile("./openapi.json")
	if err != nil {
		t.Error("Loading the openapi specification failed.")
	}
	ws := UserResource{}.WebService()
	expected := asStruct(string(raw))

	config := restfulspec.Config{
		WebServices:    []*restful.WebService{ws}, // you control what services are visible
		WebServicesURL: "http://localhost:8080",
		APIPath:        "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}

	actual := restfulspec.BuildSwagger(config)

	if reflect.DeepEqual(expected, asStruct(asJSON(actual))) != true {
		t.Errorf("Got:\n%v\nWanted:\n%v", asJSON(actual), asJSON(expected))
	}
}

func asJSON(v interface{}) string {
	data, _ := json.MarshalIndent(v, " ", " ")
	return string(data)
}

func asStruct(raw string) *spec2.Swagger {
	expected := &spec2.Swagger{}
	json.Unmarshal([]byte(raw), expected)
	return expected
}
