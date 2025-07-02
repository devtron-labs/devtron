{{- define "hpa.spec" -}}
  {{- $root := index . 0 -}}
  {{- $values := index . 1 -}}
  {{- $targetDeploymentName := index . 2 -}}
  {{- $isAdditionalHPA := index . 3 | default false -}}

{{- if $values.enabled }}
{{- if semverCompare ">=1.23-0" $root.Capabilities.KubeVersion.GitVersion }}
apiVersion: autoscaling/v2
{{- else if semverCompare ">=1.16-0" $root.Capabilities.KubeVersion.GitVersion }}
apiVersion: autoscaling/v2beta2
{{- else }}
apiVersion: autoscaling/v2beta1
{{- end }}
kind: HorizontalPodAutoscaler
metadata:
  {{- if $values.name }}
  name: {{ $values.name }}
  {{- else }}
  name: {{ $targetDeploymentName }}-hpa
  {{- end }}
  {{- if not (empty $values.annotations) }}
  annotations:
{{ toYaml $values.annotations | indent 4 }}
  {{- end }}
  labels:
    appId: {{ $root.Values.app | quote }}
    envId: {{ $root.Values.env | quote }}
    app: {{ template ".Chart.Name .name" $root }}
    chart: {{ template ".Chart.Name .chart" $root }}
    release: {{ $root.Release.Name }}  
  {{- if not (empty $values.labels) }}
{{ toYaml $values.labels | indent 4 }}
  {{- end }}
{{- if not (empty $root.Values.appLabels) }}
{{ toYaml $root.Values.appLabels | indent 4 }}
{{- end }}  
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: {{ $targetDeploymentName }}
  minReplicas: {{ $values.MinReplicas  }}
  maxReplicas: {{ $values.MaxReplicas }}
  metrics:
  {{- if not (empty $values.containerResource) }}
  {{- if $values.containerResource.enabled  }}
  {{- with $values.containerResource }}
  {{- if .TargetCPUUtilizationPercentage }}
  - type: ContainerResource  
    containerResource:
      name: cpu
      container: {{ $root.Chart.Name }}
      target:
        type: Utilization
        averageUtilization: {{ .TargetCPUUtilizationPercentage }}
  {{- end}}  
  {{- if .TargetMemoryUtilizationPercentage  }}
  - type: ContainerResource  
    containerResource:
      name: memory
      container: {{ $root.Chart.Name }}
      target:
        type: Utilization
        averageUtilization: {{ .TargetMemoryUtilizationPercentage }}
  {{- end}}
  {{- end }}            
  {{- end }}  
  {{- end }}
  {{- if $values.TargetMemoryUtilizationPercentage }}
  - type: Resource
    resource:
      name: memory
      {{- if semverCompare ">=1.16-0" $root.Capabilities.KubeVersion.GitVersion }}
      target:
        type: Utilization
        averageUtilization: {{ $values.TargetMemoryUtilizationPercentage }}
      {{- else }}
      targetAverageUtilization: {{ $values.TargetMemoryUtilizationPercentage }}
      {{- end }}
  {{- end }}
  {{- if $values.TargetCPUUtilizationPercentage }}
  - type: Resource
    resource:
      name: cpu
      {{- if semverCompare ">=1.16-0" $root.Capabilities.KubeVersion.GitVersion }}
      target:
        type: Utilization
        averageUtilization: {{ $values.TargetCPUUtilizationPercentage }}
      {{- else }}
      targetAverageUtilization: {{ $values.TargetCPUUtilizationPercentage }}
      {{- end }}
  {{- end }}
    {{- if and (not (empty $values.extraMetrics)) (semverCompare ">=1.16-0" $root.Capabilities.KubeVersion.GitVersion) }}
  {{- toYaml $values.extraMetrics | nindent 2 }}
    {{- end}}
  {{- if and (not (empty $values.behavior)) (semverCompare ">=1.18-0" $root.Capabilities.KubeVersion.GitVersion) }}
  behavior:
    {{- toYaml $values.behavior | nindent 4 }}
  {{- end }}
{{- end }}
{{- end -}}
