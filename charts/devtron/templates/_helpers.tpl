{{- define "argo-cd.selectorLabels" -}}
{{- if .name -}}
app.kubernetes.io/name: {{ include "argo-cd.name" .context }}-{{ .name }}
{{ end -}}
{{ end -}}
{{/*
Returns a secret if it already in Kubernetes, otherwise it creates
it randomly.
*/}}
{{- define "getOrGeneratePass" }}
{{- $len := (default 32 .Length) | int -}}
{{- $obj := (lookup "v1" .Kind .Namespace .Name).data -}}
{{- if $obj }}
{{- index $obj .Key -}}
{{- else if (eq (lower .Kind) "secret") -}}
{{- randAlphaNum $len | b64enc -}}
{{- else -}}
{{- randAlphaNum $len -}}
{{- end -}}
{{- end }}

{{- define "fullImage" }}
{{- $registryName := .component.registry | default .global.containerRegistry -}}
{{- $imageName := .image | default .component.image   -}}
{{- $imageTag := .tag | default .component.tag   -}}
{{- $imageDigest := .digest | default .component.digest -}}
{{- if and  $registryName $imageName $imageTag $imageDigest }}
    {{- printf "%s/%s:%s@%s" $registryName $imageName $imageTag $imageDigest -}}
{{- else  if and  $registryName $imageName $imageTag  -}}    
    {{- printf "%s/%s:%s" $registryName $imageName $imageTag  -}}
{{- else if and  $registryName $imageName $imageDigest }}
    {{- printf "%s/%s@%s" $registryName $imageName $imageDigest -}}
{{- else  }}
    {{- printf "%s/%s" $registryName $imageName  -}}
{{- end }}
{{- end -}}
