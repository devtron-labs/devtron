{{- define "keda.autoscaling.spec" -}}
  {{- $root := index . 0 -}}
  {{- $values := index . 1 -}}
  {{- $targetDeploymentName := index . 2 -}}

{{- if $values.enabled }}
apiVersion: keda.sh/v1alpha1
kind: ScaledObject
metadata:
  {{- if $values.name }}
  name: {{ $values.name }}
  {{- else }}
  name: {{ $targetDeploymentName }}-keda
  {{- end }}
  labels:
    app: {{ template ".Chart.Name .name" $root }}
    chart: {{ template ".Chart.Name .chart" $root }}
    release: {{ $root.Release.Name }}
    appId: {{ $root.Values.app | quote }}
    envId: {{ $root.Values.env | quote }}
    release: {{ $root.Release.Name }}
  {{- if not (empty $root.Values.appLabels) }}
{{ toYaml $root.Values.appLabels | indent 4 }}
  {{- end }}
  {{- if not (empty $values.labels) }}
{{ toYaml $values.labels | indent 4 }}
  {{- end }}
  {{- if not (empty $values.annotations) }}
  annotations:
{{ toYaml $values.annotations | indent 4 }}
  {{- end }}
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ $targetDeploymentName }}
{{- if not (empty $values.envSourceContainerName) }}
    envSourceContainerName: {{ $values.envSourceContainerName }}
{{- end }}
{{- if not (empty $values.pollingInterval) }}
  pollingInterval: {{ $values.pollingInterval }}
{{- end }}
{{- if not (empty $values.cooldownPeriod) }}
  cooldownPeriod: {{ $values.cooldownPeriod }}
{{- end }}
{{- if not (empty $values.idleReplicaCount) }}
  idleReplicaCount: {{ $values.idleReplicaCount }}
{{- end }}
  minReplicaCount: {{ $values.minReplicaCount }}
  maxReplicaCount: {{ $values.maxReplicaCount }}
{{- if not (empty $values.fallback) }}
  fallback: 
{{ toYaml $values.fallback | indent 4 }}
{{- end }}
{{- if not (empty $values.advanced) }}
  advanced: 
{{ toYaml $values.advanced | indent 4 }}
{{- end }}
  triggers:
{{ toYaml $values.triggers | indent 2}}
{{- if not (empty $values.authenticationRef) }}
    authenticationRef: 
{{ toYaml $values.authenticationRef | indent 6 }}
{{- end }}
{{- end }}

---
{{- if $values.triggerAuthentication.enabled }}
apiVersion: keda.sh/v1alpha1
kind: TriggerAuthentication
metadata:
  name: {{ $values.triggerAuthentication.name }}
  labels:
    app: {{ template ".Chart.Name .name" $root }}
    chart: {{ template ".Chart.Name .chart" $root }}
    release: {{ $root.Release.Name }}
    appId: {{ $root.Values.app | quote }}
    envId: {{ $root.Values.env | quote }}
    {{- if not (empty $root.Values.appLabels) }}
{{ toYaml $root.Values.appLabels | indent 4 }}
    {{- end }}        
spec:
{{ toYaml $values.triggerAuthentication.spec | indent 2 }}
{{- end }}
{{- end -}}
