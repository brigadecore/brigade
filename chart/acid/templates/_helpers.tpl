{{/* vim: set filetype=mustache */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "acid.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "acid.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "acid.gw.fullname" -}}
{{ include "acid.fullname" . | printf "%s-gw" }}
{{- end -}}
{{- define "acid.ctrl.fullname" -}}
{{ include "acid.fullname" . | printf "%s-ctrl" }}
{{- end -}}
{{- define "acid.api.fullname" -}}
{{ include "acid.fullname" . | printf "%s-api" }}
{{- end -}}
{{- define "acid.worker.fullname" -}}
{{ include "acid.fullname" . | printf "%s-wrk" }}
{{- end -}}

{{- define "acid.rbac.version" }}rbac.authorization.k8s.io/v1beta1{{ end -}}
