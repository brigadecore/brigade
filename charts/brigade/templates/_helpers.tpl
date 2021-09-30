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

{{- define "brigade.scheduler.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-scheduler" }}
{{- end -}}

{{- define "brigade.observer.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-observer" }}
{{- end -}}

{{- define "brigade.artemis.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-artemis" }}
{{- end -}}

{{- define "brigade.logger.fullname" -}}
{{ include "brigade.fullname" . | printf "%s-logger" }}
{{- end -}}

{{- define "brigade.logger.linux.fullname" -}}
{{ include "brigade.logger.fullname" . | printf "%s-linux" }}
{{- end -}}

{{- define "brigade.logger.windows.fullname" -}}
{{ include "brigade.logger.fullname" . | printf "%s-windows" }}
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

{{- define "brigade.scheduler.labels" -}}
app.kubernetes.io/component: scheduler
{{- end -}}

{{- define "brigade.observer.labels" -}}
app.kubernetes.io/component: observer
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

{{- define "brigade.logger.labels" -}}
app.kubernetes.io/component: logger
{{- end -}}

{{- define "brigade.logger.linux.labels" -}}
{{ include "brigade.logger.labels" . }}
app.kubernetes.io/os: linux
{{- end -}}

{{- define "brigade.logger.windows.labels" -}}
{{ include "brigade.logger.labels" . }}
app.kubernetes.io/os: windows
{{- end -}}

{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end }}

{{/*
Return the appropriate apiVersion for a networking object.
*/}}
{{- define "networking.apiVersion" -}}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- end -}}
{{- end -}}

{{- define "networking.apiVersion.isStable" -}}
  {{- eq (include "networking.apiVersion" .) "networking.k8s.io/v1" -}}
{{- end -}}

{{- define "networking.apiVersion.supportIngressClassName" -}}
  {{- semverCompare ">=1.18-0" .Capabilities.KubeVersion.GitVersion -}}
{{- end -}}
