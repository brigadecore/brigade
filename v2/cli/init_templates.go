package main

import (
	"bytes"
	"text/template"

	"github.com/Masterminds/sprig"
)

// Template project.yaml file for Brigade projects
// nolint: lll
var projectTemplate = []byte(`# yaml-language-server: $schema=https://raw.githubusercontent.com/brigadecore/brigade/v2/v2/apiserver/schemas/project.json
apiVersion: brigade.sh/v2-beta
kind: Project
metadata:
  id: {{ .ProjectID }}
description: My new Brigade project
spec:
  ## Subscribe below to any events that should trigger your script.
  ## The example depicts a subscription to "exec" events 
  ## originating from the Brigade CLI. 
  eventSubscriptions:
  - source: brigade.sh/cli
    types:
    - exec
  workerTemplate:
    logLevel: DEBUG
    {{- if eq .GitCloneURL "" }}
    defaultConfigFiles:
      brigade.{{ .Language }}: |
        {{- nindent 8 .Script }}
    {{- else }}
    ## Your brigade.{{ .Language }} script will be loaded from the 
    ## following git repository by default, but note that individual 
    ## events CAN override these details. The Brigade GitHub Gateway, 
    ## for instance, overrides these details for relevant events so 
    ## the events themselves reference applicable branches, tags, pull 
    ## requests, etc.
    git:
      cloneURL: {{ .GitCloneURL }}
    {{- end }}
`)

// nolint: lll
var typeScriptTemplate = []byte(`import { events, Job } from "@brigadecore/brigadier"

// Use events.on() to define how your script responds to different events. 
// The example below depicts handling of "exec" events originating from
// the Brigade CLI.

events.on("brigade.sh/cli", "exec", async event => {
	let job = new Job("hello", "debian:latest", event)
 	job.primaryContainer.command = ["echo"]
 	job.primaryContainer.arguments = ["Hello, World!"]
 	await job.run()
})

events.process()
`)

// nolint: lll
var javaScriptTemplate = []byte(`const { events } = require("@brigadecore/brigadier");

// Use events.on() to define how your script responds to different events. 
// The example below depicts handling of "exec" events originating from
// the Brigade CLI.

events.on("brigade.sh/cli", "exec", async event => {
 	let job = new Job("hello", "debian:latest", event);
 	job.primaryContainer.command = ["echo"];
 	job.primaryContainer.arguments = ["Hello, World!"];
 	await job.run();
});

events.process();
`)

// nolint: lll
var notesTemplate = []byte(`This file was created by brig init and outlines next steps.

1. Edit project.yaml as you see fit and submit it to Brigade using:

     brig project create -f .brigade/project.yaml

2. If your project needs to make use of any secret/sensitive values, set them
   one at a time using:

     brig project secrets set --id {{ .ProjectID }} --set <key>=<value>

   Or set secrets in bulk by modifying secrets.yaml, then submit them to
   Brigade using:

     brig project secrets set --id {{ .ProjectID }} --file .brigade/secrets.yaml

   Secrets become available within your brigade.{{ .Language }} script using:

     event.project.secrets.<key>

   Take great care in how you use secrets so they are not leaked into worker or
   job logs.

   Note that Brigade keeps all jobs' environment variables secret, so it is safe
   to inject your secrets into a job's environment. This action will not
   implicitly expose your secrets. Take care that the jobs themselves don't leak
   secrets into their logs.

   DO NOT commit secrets to source control. brig init has already added
   .brigade/secrets.yaml to your .gitignore file.

{{- if not (eq .GitCloneURL "") }}

3. Since your project expects to handle events using your brigade.{{ .Language}}
    script found at git repository {{ .GitCloneURL }},
    be sure to commit and push your brigade.{{ .Language }} script to that
    location.
{{- end }}
`)

var secretsTemplate = []byte(`## This file was created by brig init.
##
## Specify values for project secrets using key/value pairs below.
##
## You can load all these secrets into your project using
##
##   brig project secrets set --id {{ .ProjectID }} --file .brigade/secrets.yaml
##
## DO NOT commit these secrets to source control. brig init has already added
## .brigade/secrets.yaml to your .gitignore file.
##
## Example keys/values:
##
# foo: bar
# bat: baz
`)

var packageTemplate = []byte(`{
  "name": "{{ .ProjectID }}",
  "dependencies": {
    "@brigadecore/brigadier": "^2.0.0-beta.2"
  }
}
`)

// execTemplate applies the given fields to the given template and returns the
// template bytes.
func execTemplate(
	templateBytes []byte,
	fields interface{},
) ([]byte, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New(
		"template",
	).Funcs(sprig.TxtFuncMap()).Parse(string(templateBytes))
	if err != nil {
		return buf.Bytes(), err
	}

	if err = tmpl.Execute(buf, fields); err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
