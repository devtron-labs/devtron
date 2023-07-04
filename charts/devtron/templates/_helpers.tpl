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
{{/*
Validate data (security)
*/}}
{{- define "devtron.validateValues" }}
{{- if .Values.security }}
{{- if and (.Values.security.clair.enabled) (.Values.security.trivy.enabled) }}
{{- printf "Enable either clair or trivy, enabling both at the same time isn't supported as of now" | fail }}
{{- end }}
{{- end }}
{{- end }}
