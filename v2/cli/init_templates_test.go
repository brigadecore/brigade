package main

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestExecTemplate(t *testing.T) {
	bytes, err := execTemplate(
		[]byte(`hello {{.ID}}`),
		struct{ ID string }{ID: "world"},
	)

	require.NoError(t, err)

	require.Equal(t, []byte(`hello world`), bytes)
}

func TestProjectFileTemplate(t *testing.T) {
	testCases := []struct {
		name        string
		projectID   string
		language    string
		gitCloneURL string
	}{
		{name: "nonGit", projectID: "hello", language: "ts", gitCloneURL: ""},
		{name: "git", projectID: "hello", language: "ts",
			gitCloneURL: "https://github.com/brigadecore/brigade.git"},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			bytes, err :=
				execTemplate(projectTemplate, struct {
					ProjectID   string
					Language    string
					GitCloneURL string
					Script      string
				}{
					ProjectID:   testCase.projectID,
					Language:    testCase.language,
					GitCloneURL: testCase.gitCloneURL,
					Script:      "",
				})

			require.NoError(t, err)

			results := map[string]interface{}{}
			err = yaml.Unmarshal(bytes, &results)
			require.NoError(t, err)

			jsonBytes, err := yaml.YAMLToJSON(bytes)
			require.NoError(t, err)

			loader := gojsonschema.NewReferenceLoader(
				"file://../apiserver/schemas/project.json",
			)

			validationResult, err := gojsonschema.Validate(
				loader,
				gojsonschema.NewBytesLoader(jsonBytes),
			)

			require.NoError(t, err)
			require.True(t, validationResult.Valid())
		})
	}
}
