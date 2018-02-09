{{/* vim: set filetype=mustache */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "brigade.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "brigade.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "brigade.gw.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-github-gw" }}
{{- end -}}
{{- define "brigade.ctrl.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-ctrl" }}
{{- end -}}
{{- define "brigade.api.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-api" }}
{{- end -}}
{{- define "brigade.worker.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-wrk" }}
{{- end -}}
{{- define "brigade.cr.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-cr-gw" }}
{{- end -}}
{{- define "brigade.vacuum.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-vacuum" }}
{{- end -}}

{{- define "brigade.rbac.version" }}rbac.authorization.k8s.io/v1beta1{{ end -}}
