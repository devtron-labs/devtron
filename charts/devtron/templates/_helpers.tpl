{{/*
Returns external database password
*/}}
{{- define "pg.password" }}
{{- if .Values.global.externalDatabase }}
{{- $secretData := .Values.global.externalDatabase.PG_PASSWORD }}
{{- $Secret := $secretData | b64enc }}
{{- $Secret }}
{{- else }}
{{- $secretObj := (lookup "v1" "Secret" "devtroncd" "postgresql-postgresql") | default dict }}
{{- $secretData := (get $secretObj "data") | default dict }}
{{- $Secret := (get $secretData "postgresql-password") | default (randAlphaNum 32) | b64enc }}
{{- $Secret }}
{{- end }}
{{- end }}

{{/*
Returns a service name for Clair
*/}}
{{- define "clair.service" }}
{{- if .Values.global.externalDatabase }}
{{- print "clair-postgresql" }}
{{- else }}
{{- print "postgresql-postgresql" }}
{{- end }}
{{- end }}

{{/*
Returns a db name, clairv4 or orchestrator 
*/}}
{{- define "clair.db" }}
{{- if .Values.global.externalDatabase }}
{{- print "clairv4" }}
{{- else }}
{{- print "orchestrator" }}
{{- end }}
{{- end }}

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
