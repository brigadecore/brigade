{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "brigade.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "brigade.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "brigade.apiserver.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-apiserver" }}
{{- end -}}

{{- define "brigade.artemis.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-artemis" }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "brigade.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "brigade.labels" -}}
helm.sh/chart: {{ include "brigade.chart" . }}
{{ include "brigade.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "brigade.selectorLabels" -}}
app.kubernetes.io/name: {{ include "brigade.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "brigade.apiserver.labels" -}}
app.kubernetes.io/component: apiserver
{{- end -}}

{{- define "brigade.artemis.labels" -}}
app.kubernetes.io/component: artemis
{{- end -}}

{{- define "brigade.artemis.primary.labels" -}}
{{ include "brigade.artemis.labels" . }}
app.kubernetes.io/role: primary
{{- end -}}

{{- define "brigade.artemis.secondary.labels" -}}
{{ include "brigade.artemis.labels" . }}
app.kubernetes.io/role: secondary
{{- end -}}

{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end }}
