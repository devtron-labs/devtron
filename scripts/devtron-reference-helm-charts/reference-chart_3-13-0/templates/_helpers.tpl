{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define ".Chart.Name .name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create service name
*/}}
{{- define ".servicename" -}}
{{- if .Values.service.name -}}
{{- .Values.service.name | trunc 63 | trimSuffix "-" -}}
{{- else if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 55 | trimSuffix "-" -}}-service
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 55 | trimSuffix "-" -}}-service
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 55 | trimSuffix "-" -}}-service
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create preview service name
*/}}
{{- define ".previewservicename" -}}
{{- if .Values.service.name -}}
{{- .Values.service.name | trunc 55 | trimSuffix "-" -}}-preview
{{- else if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 47 | trimSuffix "-" -}}-preview-service
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 47 | trimSuffix "-" -}}-preview-service
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 47 | trimSuffix "-" -}}-preview-service
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define ".Chart.Name .fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define ".Chart.Name .chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define ".Chart.Name .color" -}}
{{- $active0 := (index .Values.server.deployment 0).enabled -}}
{{/*
{{- $active1 := (index .Values.server.deployment 1).enabled -}}
*/}}
{{- $active1 := include "safeenabledcheck" . -}}
{{- $active := and $active0 $active1 -}}
{{- $active -}}
{{- end -}}

{{- define "safeenabledcheck" -}}
{{- if (eq (len .Values.server.deployment) 2) -}}
  {{- if (index .Values.server.deployment 1).enabled -}}
    {{- $active := true -}}
    {{- $active -}}
  {{- else -}}
    {{-  $active := false -}}
    {{- $active -}}
  {{- end -}}
{{- else -}}
  {{- $active := false -}}
  {{- $active -}}
{{- end -}}
{{- end -}}


{{- define "isCMVolumeExists" -}}
  {{- $isCMVolumeExists := false -}}
    {{- if .Values.ConfigMaps.enabled }}
      {{- range .Values.ConfigMaps.maps }}
        {{- if eq .type "volume"}}
          {{- $isCMVolumeExists = true}}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- $isCMVolumeExists -}}
{{- end -}}

{{- define "isSecretVolumeExists" -}}
  {{- $isSecretVolumeExists := false -}}
    {{- if .Values.ConfigSecrets.enabled }}
      {{- range .Values.ConfigSecrets.secrets }}
        {{- if eq .type "volume"}}
          {{- $isSecretVolumeExists = true}}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- $isSecretVolumeExists -}}
{{- end -}}




{{- define "serviceMonitorEnabled" -}}
   {{- $SMenabled := false -}}
   {{- range .Values.ContainerPort }}
       {{- if .servicemonitor }}
             {{- if and .servicemonitor.enabled }}
                 {{- $SMenabled = true -}}
             {{- end }}
       {{- end }}
   {{- end }}
   {{- $SMenabled -}}
{{- end -}}
