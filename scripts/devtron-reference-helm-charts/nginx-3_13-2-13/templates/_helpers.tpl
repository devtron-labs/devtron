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

{{/* vim: set filetype=mustache: */}}
{{/*
Return the proper NGINX image name
*/}}
{{- define "nginx.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper GIT image name
*/}}
{{- define "nginx.cloneStaticSiteFromGit.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.cloneStaticSiteFromGit.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Prometheus metrics image name
*/}}
{{- define "nginx.metrics.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.metrics.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "nginx.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image .Values.cloneStaticSiteFromGit.image .Values.metrics.image) "global" .Values.global) }}
{{- end -}}

{{/*
Return true if a static site should be mounted in the NGINX container
*/}}
{{- define "nginx.useStaticSite" -}}
{{- if or .Values.cloneStaticSiteFromGit.enabled .Values.staticSiteConfigmap .Values.staticSitePVC }}
    {- true -}}
{{- end -}}
{{- end -}}

{{/*
Return the volume to use to mount the static site in the NGINX container
*/}}
{{- define "nginx.staticSiteVolume" -}}
{{- if .Values.cloneStaticSiteFromGit.enabled }}
emptyDir: {}
{{- else if .Values.staticSiteConfigmap }}
configMap:
  name: {{ printf "%s" (tpl .Values.staticSiteConfigmap $) -}}
{{- else if .Values.staticSitePVC }}
persistentVolumeClaim:
  claimName: {{ printf "%s" (tpl .Values.staticSitePVC $) -}}
{{- end }}
{{- end -}}

{{/*
Return the custom NGINX server block configmap.
*/}}
{{- define "nginx.serverBlockConfigmapName" -}}
{{- if .Values.existingServerBlockConfigmap -}}
    {{- printf "%s" (tpl .Values.existingServerBlockConfigmap $) -}}
{{- else -}}
    {{- printf "%s-server-block" (include "common.names.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Compile all warnings into a single message, and call fail.
*/}}
{{- define "nginx.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "nginx.validateValues.cloneStaticSiteFromGit" .) -}}
{{- $messages := append $messages (include "nginx.validateValues.extraVolumes" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message | fail -}}
{{- end -}}
{{- end -}}

{{/* Validate values of NGINX - Clone StaticSite from Git configuration */}}
{{- define "nginx.validateValues.cloneStaticSiteFromGit" -}}
{{- if and .Values.cloneStaticSiteFromGit.enabled (or (not .Values.cloneStaticSiteFromGit.repository) (not .Values.cloneStaticSiteFromGit.branch)) -}}
nginx: cloneStaticSiteFromGit
    When enabling cloing a static site from a Git repository, both the Git repository and the Git branch must be provided.
    Please provide them by setting the `cloneStaticSiteFromGit.repository` and `cloneStaticSiteFromGit.branch` parameters.
{{- end -}}
{{- end -}}

{{/* Validate values of NGINX - Incorrect extra volume settings */}}
{{- define "nginx.validateValues.extraVolumes" -}}
{{- if and (.Values.extraVolumes) (not (or .Values.extraVolumeMounts .Values.cloneStaticSiteFromGit.extraVolumeMounts)) -}}
nginx: missing-extra-volume-mounts
    You specified extra volumes but not mount points for them. Please set
    the extraVolumeMounts value
{{- end -}}
{{- end -}}
{{/* Create the name of the service account to use */}}
{{- define "serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include ".Chart.Name .fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

