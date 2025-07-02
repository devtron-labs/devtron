{{- define "deployment.spec" -}}
  {{- $root := index . 0 -}}
  {{- $values := index . 1 -}}
  {{- $nameSuffix := index . 2 | default "" -}}
  {{- $isAdditionalDeployment := index . 3 | default false -}}

  {{- $hasCMEnvExists := false -}}
  {{- $hasCMVolumeExists := false -}}
  {{- if $values.ConfigMaps.enabled }}
  {{- range $values.ConfigMaps.maps | default (list) }}
  {{- if eq .type "volume"}}
  {{- $hasCMVolumeExists = true}}
  {{- end }}
  {{- if eq .type "environment"}}
  {{- $hasCMEnvExists = true}}
  {{- end }}
  {{- end }}
  {{- end }}

  {{- $hasPVCExists := false -}}
  {{- if $values.persistentVolumeClaim.name }}
  {{- $hasPVCExists = true }}
  {{- end }}

  {{- $hasSecretEnvExists := false -}}
  {{- $hasSecretVolumeExists := false -}}
  {{- if $values.ConfigSecrets.enabled }}
  {{- range $values.ConfigSecrets.secrets | default (list) }}
  {{- if eq .type "volume"}}
  {{- $hasSecretVolumeExists = true}}
  {{- end }}
  {{- if eq .type "environment"}}
  {{- $hasSecretEnvExists = true}}
  {{- end }}
  {{- end }}
  {{- end }}
  {{ $CustomLabelsApp:= include "customPodLabelsContainsApp" $root }}
  {{ $CustomLabelsRelease:= include "customPodLabelsContainsRelease" $root }}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ if $values.nameOverride }}{{ $values.nameOverride }}{{ else }}{{ include ".Chart.Name .fullname" $root }}{{ if $isAdditionalDeployment }}-{{ $nameSuffix }}{{ end }}{{ end }}
  labels:
    app: {{ template ".Chart.Name .name" $root }}
    chart: {{ template ".Chart.Name .chart" $root }}
    release: {{ $root.Release.Name }}
    releaseVersion: {{ $values.releaseVersion | quote }}
    pipelineName: {{ $values.pipelineName }}
{{- if not (empty $values.deploymentLabels) }}
{{ toYaml $values.deploymentLabels | indent 4 }}
{{- end }}
{{- if not (empty $values.appLabels) }}
{{ toYaml $values.appLabels | indent 4 }}
{{- end }}

{{- if not (empty $values.deploymentAnnotations) }}
  annotations:
{{ toYaml $values.deploymentAnnotations | indent 4 }}
{{- end }}
spec:
  selector:
    matchLabels:
{{- if not (empty $values.customMatchLabels)  }}
{{ toYaml $values.customMatchLabels | indent 6 }}  
{{- else }}
      app: {{ $values.customPodLabels.app | default  (include ".Chart.Name .name" $root) }}
      release: {{ $values.customPodLabels.release | default  $root.Release.Name }}
{{- end }}
  replicas: {{ $values.replicaCount }}
  minReadySeconds: {{ $values.MinReadySeconds }}
  template:
    metadata:
    {{- if not (empty $values.podAnnotations) }}
      annotations:
      {{- range $key, $value := $values.podAnnotations }}
        {{ $key }}: {{ $value | quote }}
      {{- end }}
    {{- end }}
      labels:
        {{- if not (eq "true" $CustomLabelsApp) }}
        app: {{ $values.customPodLabels.app | default  (include ".Chart.Name .name" $root) }}
        {{- end }}
        {{- if not (eq "true" $CustomLabelsRelease) }}
        release: {{ $values.customPodLabels.release |default  $root.Release.Name }}
        {{- end }}
        appId: {{ $root.Values.app | quote }}
        envId: {{ $root.Values.env | quote }}
{{- if not (empty $values.customPodLabels)  }}
{{ toYaml $values.customPodLabels | indent 8 }}  
{{- end }}                
{{- if not (empty $values.appLabels) }}
{{ toYaml $values.appLabels | indent 8 }}
{{- end }}
{{- if not (empty $values.podLabels) }}
{{ toYaml $values.podLabels | indent 8 }}
{{- end }}
    spec:
{{- if not (empty $values.podExtraSpecs) }}	
{{ toYaml $values.podExtraSpecs | indent 6 }}	
{{- end }}
      terminationGracePeriodSeconds: {{ $values.GracePeriod }}
      restartPolicy: Always
{{- if not (empty $values.hostAliases) }}
      hostAliases:
{{ toYaml $values.hostAliases | indent 8 }}
{{- end }}
{{- if or (and $values.Spec.Affinity.Key $values.Spec.Affinity.Values) ($values.affinity.enabled) }}
      affinity:
{{- if and $values.Spec.Affinity.Key $values.Spec.Affinity.Values }}
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: {{ $values.Spec.Affinity.Key  }}
                operator: In
                values:
                - {{ $values.Spec.Affinity.Values | default "nodes"  }}
{{- else if $values.affinity.enabled }}
      {{- if not (empty $values.affinity.values) }}
        {{ toYaml $values.affinity.values | indent 8 }}
      {{- end }}
{{- end }}
{{- end }}
{{- if $values.serviceAccountName }}
      serviceAccountName: {{ $values.serviceAccountName }}
{{- else  }}
      serviceAccountName: {{ template "serviceAccountName" $root }}
{{- end }}
{{- if $values.schedulerName }}
      schedulerName: {{ $values.schedulerName }}
{{- end }}      
  {{- if not (empty $values.tolerations) }}
      tolerations:
{{ toYaml $values.tolerations | indent 8 }}
  {{- end }}
{{- if not (empty $values.imagePullSecrets) }}
      imagePullSecrets:
  {{- range $values.imagePullSecrets }}
        - name: {{ . }}
  {{- end }}
{{- end}}
{{- if not (empty $values.topologySpreadConstraints) }}
      topologySpreadConstraints:
{{- range $values.topologySpreadConstraints }}
      - maxSkew: {{ .maxSkew }}
        topologyKey: {{ .topologyKey }}
        whenUnsatisfiable: {{ .whenUnsatisfiable }}
        {{- if semverCompare "<=1.30-0" $root.Capabilities.KubeVersion.GitVersion }}
        {{- if .minDomains }}
        minDomains: {{ .minDomains }}
        {{- end }}
        {{- end }}        
        {{- if .nodeAffinityPolicy }}
        nodeAffinityPolicy: {{ .nodeAffinityPolicy }}
        {{- end }}
        {{- if .nodeTaintsPolicy }}
        nodeTaintsPolicy: {{ .nodeTaintsPolicy }}
        {{- end }}                 
        labelSelector:
          matchLabels:
          {{- if and .autoLabelSelector .customLabelSelector }}
{{ toYaml .customLabelSelector | indent 12 }}
          {{- else if .autoLabelSelector }}
            app: {{ template ".Chart.Name .name" $root }}
            appId: {{ $root.Values.app | quote }}
            envId: {{ $root.Values.env | quote }}
            release: {{ $root.Release.Name }}
          {{- else if .customLabelSelector }}
{{ toYaml .customLabelSelector | indent 12 }}
          {{- end }}
{{- end }}
{{- end }}
{{- if not (empty $values.topologySpreadConstraint) }}
      topologySpreadConstraints: 
{{ toYaml $values.topologySpreadConstraint }}  
{{- end }}    
{{- if not (empty $values.podSecurityContext) }}
      securityContext:
{{ toYaml $values.podSecurityContext | indent 8 }}
{{- end }}
{{- if not (empty $values.restartPolicy) }}
      restartPolicy: {{  $values.restartPolicy }}
{{- end }}
{{- if not (empty $values.initContainers) }}
      initContainers:
{{- range $i, $c := $values.initContainers }}
{{- if .reuseContainerImage}}
        - name: {{ $root.Chart.Name }}-init-{{ add1 $i }}
          image: "{{ $values.server.deployment.image }}:{{ $values.server.deployment.image_tag }}"
          imagePullPolicy: {{ $values.image.pullPolicy }}
{{- if not (empty .securityContext) }}
          securityContext:
{{ toYaml .securityContext | indent 12 }}
{{- end }}
{{- if not (empty .command) }}
          command:
{{ toYaml .command | indent 12 -}}
{{- end}}
{{- if not (empty .args) }}
          args:
{{ toYaml .args | indent 12 -}}
{{- end}}
{{- if not (empty .resources) }}
          resources:
{{ toYaml .resources | indent 12 -}}
{{- end}}
{{- if not (empty .volumeMounts) }}
          volumeMounts:
{{ toYaml .volumeMounts | indent 12 -}}
{{- end}}
{{- else}}
        -
{{ toYaml . | indent 10 }}
{{- end}}
{{- end}}
{{- end}}
      containers:
        - name: {{ $root.Chart.Name }}
          image: "{{ $values.server.deployment.image }}:{{ $values.server.deployment.image_tag }}"
          imagePullPolicy: {{ $values.image.pullPolicy }}
          {{- if and $values.containerSpec.lifecycle.enabled (or $values.containerSpec.lifecycle.preStop (not (empty $values.containerSpec.lifecycle.postStart))) }}
          lifecycle:
           {{- if not (empty $values.containerSpec.lifecycle.preStop) }}
           preStop:
{{ toYaml $values.containerSpec.lifecycle.preStop | indent 12 -}}
           {{- end }}
           {{- if not (empty $values.containerSpec.lifecycle.postStart) }}
           postStart:
{{ toYaml $values.containerSpec.lifecycle.postStart | indent 12 -}}
           {{- end }}
          {{- end }}
{{- if not (empty $values.containerSecurityContext) }}
          securityContext:
{{- if $values.privileged }}
            privileged: true
{{- end }}
{{ toYaml $values.containerSecurityContext | indent 12 }}
{{- else if $values.privileged }}
          securityContext:
            privileged: true
{{- end }}
{{- if not (empty $values.containerExtraSpecs) }}	
{{ toYaml $values.containerExtraSpecs | indent 10 }}	
{{- end }}
{{- if not (empty $values.resizePolicy) }}
          resizePolicy:
{{ toYaml $values.resizePolicy | indent 12 }}          
{{- end }}          
          ports:
          {{- if not (empty $values.ContainerPort) }}
          {{- range $values.ContainerPort }}
            - name: {{ .name}}
              containerPort: {{ .port  }}
              protocol: {{ .protocol | default "TCP" }}
          {{- end}}
          {{- end }}
{{- if and $values.command.enabled (not (empty $values.command.workingDir)) }}
          workingDir: {{ $values.command.workingDir }}
{{- end}}
{{- if and $values.command.enabled (not (empty $values.command.value)) }}
          command:
{{ toYaml $values.command.value | indent 12 -}}
{{- end}}
{{- if and $values.args.enabled (not (empty $values.args.value)) }}
          args:
{{ toYaml $values.args.value | indent 12 -}}
{{- end }}
          env:
            - name: CONFIG_HASH
              value: {{ include (print $root.Template.BasePath "/configmap.yaml") $root | sha256sum }}{{ if and ($values.devtronInternal) ($values.devtronInternal.containerSpecs.ConfigHash) }}{{ $values.devtronInternal.containerSpecs.ConfigHash }}{{ end }}
            - name: SECRET_HASH
              value: {{ include (print $root.Template.BasePath "/secret.yaml") $root | sha256sum }}{{ if and ($values.devtronInternal) ($values.devtronInternal.containerSpecs.SecretHash) }}{{ $values.devtronInternal.containerSpecs.SecretHash }}{{ end }}
            - name: DEVTRON_APP_NAME
              value: {{ template ".Chart.Name .name" $root }}
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: DEVTRON_CONTAINER_REPO
              value: "{{ $values.server.deployment.image }}"
            - name: DEVTRON_CONTAINER_TAG
              value: "{{ $values.server.deployment.image_tag }}"
          {{- if not (empty $values.EnvVariablesFromFieldPath) }}
          {{- range $values.EnvVariablesFromFieldPath }}
            - name: {{ .name }}
              valueFrom:
                fieldRef:
                 fieldPath: {{ .fieldPath }}
          {{- end}}
          {{- end }}
          {{- if not (empty $values.EnvVariables) }}
          {{- range $values.EnvVariables }}
          {{- if and .name .value }}
            - name: {{ .name }}
              value: {{ .value | quote }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if not (empty $values.EnvVariablesFromSecretKeys) }}
          {{- range $values.EnvVariablesFromSecretKeys }}
          {{- if and .name .secretName .keyName }}
            - name: {{ .name }}
              valueFrom:
                secretKeyRef:
                  name: {{ .secretName }}
                  key: {{ .keyName }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if not (empty $values.EnvVariablesFromConfigMapKeys) }}
          {{- range $values.EnvVariablesFromConfigMapKeys }}
          {{- if and .name .configMapName .keyName }}
            - name: {{ .name }}
              valueFrom:
                configMapKeyRef:
                  name: {{ .configMapName }}
                  key: {{ .keyName }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if or (and ($hasCMEnvExists) ($values.ConfigMaps.enabled)) (and ($hasSecretEnvExists) ($values.ConfigSecrets.enabled)) }}
          envFrom:
          {{- if $values.ConfigMaps.enabled }}
          {{- range $values.ConfigMaps.maps | default (list) }}
          {{- if eq .type "environment" }}
          - configMapRef:
              {{- if eq .external true }}
              name: {{ .name }}
              {{- else if eq .external false }}
              name: {{ .name}}-{{ $root.Values.app }}
              {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- if $values.ConfigSecrets.enabled }}
          {{- range $values.ConfigSecrets.secrets | default (list) }}
          {{- if eq .type "environment" }}
          - secretRef:
              {{if eq .external true}}
              name: {{ .name }}
              {{else if eq .external false}}
              name: {{ .name}}-{{ $root.Values.app }}
              {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}

{{- if or (not (empty $values.LivenessProbe.Path)) (not (empty $values.LivenessProbe.command)) (not (empty $values.LivenessProbe.tcp)) (not (empty $values.LivenessProbe.grpc)) }}
          livenessProbe:
{{- if not (empty $values.LivenessProbe.Path) }}
            httpGet:
              path: {{ $values.LivenessProbe.Path  }}
              port: {{ $values.LivenessProbe.port }}
              scheme: {{ $values.LivenessProbe.scheme }}
            {{- if not (empty $values.LivenessProbe.httpHeaders) }}
              httpHeaders:
              {{- range $values.LivenessProbe.httpHeaders}}
                - name: {{.name}}
                  value: {{.value}}
              {{- end}}
	    {{- end }}
{{- end }}
{{- if not (empty $values.LivenessProbe.command) }}
            exec:
              command:
{{ toYaml $values.LivenessProbe.command | indent 16 }}
{{- end}}
{{- if not (empty $values.LivenessProbe.tcp) }}
            tcpSocket:
              port: {{ $values.LivenessProbe.port }}
{{- end}}
            initialDelaySeconds: {{ $values.LivenessProbe.initialDelaySeconds  }}
            periodSeconds: {{ $values.LivenessProbe.periodSeconds  }}
            successThreshold: {{ $values.LivenessProbe.successThreshold  }}
            timeoutSeconds: {{ $values.LivenessProbe.timeoutSeconds  }}
            failureThreshold: {{ $values.LivenessProbe.failureThreshold  }}
          {{- if not (empty $values.LivenessProbe.grpc) }}
            grpc:
{{ toYaml $values.LivenessProbe.grpc | indent 14 }}
          {{- end }}
{{- end }}
{{- if or (not (empty $values.ReadinessProbe.Path)) (not (empty $values.ReadinessProbe.command)) (not (empty $values.ReadinessProbe.tcp)) (not (empty $values.ReadinessProbe.grpc)) }}
          readinessProbe:
{{- if not (empty $values.ReadinessProbe.Path) }}
            httpGet:
              path: {{ $values.ReadinessProbe.Path  }}
              port: {{ $values.ReadinessProbe.port }}
              scheme: {{ $values.ReadinessProbe.scheme }}
            {{- if not (empty $values.ReadinessProbe.httpHeaders) }}
              httpHeaders:
              {{- range $values.ReadinessProbe.httpHeaders}}
                - name: {{.name}}
                  value: {{.value}}
              {{- end}}
	    {{- end }}
{{- end }}
{{- if not (empty $values.ReadinessProbe.command) }}
            exec:
              command:
{{ toYaml $values.ReadinessProbe.command | indent 16 }}
{{- end}}
{{- if not (empty $values.ReadinessProbe.tcp) }}
            tcpSocket:
              port: {{ $values.ReadinessProbe.port }}
{{- end}}
            initialDelaySeconds: {{ $values.ReadinessProbe.initialDelaySeconds  }}
            periodSeconds: {{ $values.ReadinessProbe.periodSeconds  }}
            successThreshold: {{ $values.ReadinessProbe.successThreshold  }}
            timeoutSeconds: {{ $values.ReadinessProbe.timeoutSeconds  }}
            failureThreshold: {{ $values.ReadinessProbe.failureThreshold  }}
          {{- if not (empty $values.ReadinessProbe.grpc) }}
            grpc:
{{ toYaml $values.ReadinessProbe.grpc | indent 14 }}
            {{- end}}
{{- end }}
          resources:
{{ toYaml $values.resources | trim | indent 12 }}
{{- if or (not (empty $values.StartupProbe.Path)) (not (empty $values.StartupProbe.command)) (not (empty $values.StartupProbe.tcp)) (not (empty $values.StartupProbe.grpc)) }}
          startupProbe:
{{- if not (empty $values.StartupProbe.Path) }}
            httpGet:
              path: {{ $values.StartupProbe.Path  }}
              port: {{ $values.StartupProbe.port }}
            {{- if not (empty $values.StartupProbe.httpHeaders) }}
              httpHeaders:
              {{- range $values.StartupProbe.httpHeaders}}
                - name: {{.name}}
                  value: {{.value}}
              {{- end}}
        {{- end }}
{{- end }}
{{- if not (empty $values.StartupProbe.command) }}
            exec:
              command:
{{ toYaml $values.StartupProbe.command | indent 16 }}
{{- end}}
{{- if not (empty $values.StartupProbe.tcp) }}
            tcpSocket:
              port: {{ $values.StartupProbe.port }}
{{- end}}
            initialDelaySeconds: {{ $values.StartupProbe.initialDelaySeconds  }}
            periodSeconds: {{ $values.StartupProbe.periodSeconds  }}
            successThreshold: {{ $values.StartupProbe.successThreshold  }}
            timeoutSeconds: {{ $values.StartupProbe.timeoutSeconds  }}
            failureThreshold: {{ $values.StartupProbe.failureThreshold  }}
            {{- if not (empty $values.StartupProbe.grpc) }}
            grpc:
{{ toYaml $values.StartupProbe.grpc | indent 14 }}
            {{- end}}
{{- end }}
          volumeMounts:
{{- if not (empty $values.volumeMounts) }}
{{ toYaml $values.volumeMounts | trim | indent 12 }}
{{- end }}
{{- if  $values.persistentVolumeClaim.name }}
            - name: {{ $values.persistentVolumeClaim.name }}-vol
              mountPath: {{ $values.persistentVolumeClaim.mountPath | default "/tmp" }}
{{- end}}
          {{- if $values.ConfigMaps.enabled }}
          {{- range $values.ConfigMaps.maps | default (list) }}
          {{- if eq .type "volume"}}
          {{- $cmName := .name -}}
          {{- $cmMountPath := .mountPath -}}
          {{- if eq .subPath false }}
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath }}
	  
          {{- else }}
          {{- range $k, $v := .data }}
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath }}/{{ $k}}
              subPath: {{ $k}}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}

          {{- if $values.ConfigSecrets.enabled }}
          {{- range $values.ConfigSecrets.secrets | default (list) }}
          {{- if eq .type "volume"}}
          {{- $cmName := .name -}}
          {{- $cmMountPath := .mountPath -}}
          {{- if eq .subPath false }}
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath }}
	      
          {{- else }}
          {{if (or (eq .externalType "ESO_GoogleSecretsManager") (eq .externalType "ESO_AWSSecretsManager") (eq .externalType "ESO_HashiCorpVault") (eq .externalType "ESO_AzureSecretsManager"))}}
          {{- if and (.esoSubPath) (ne (len .esoSubPath) 0) }}
          {{- range .esoSubPath }}
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath}}/{{ . }}
              subPath: {{ . }}  
          {{- end }}
          {{- else }}
          {{- range .esoSecretData.esoData }}
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath}}/{{ .secretKey }}
              subPath: {{ .secretKey }}  
          {{- end }}
          {{- end }}
          {{- else }}            
          {{- range $k, $v := .data }} # for others secrets the mount path will be .data[i].secretKey
            - name: {{ $cmName | replace "." "-"}}-vol
              mountPath: {{ $cmMountPath}}/{{ $k}}
              subPath: {{ $k}}
          {{- end }}
          {{- end }}          
          {{- end }}
          {{- end }}
          {{- end }}
          {{- end }}

{{- if $values.appMetrics }}
        - name: envoy
          image: {{ $values.envoyproxy.image | default "quay.io/devtron/envoy:v1.16.0"}}
          {{- if $values.envoyproxy.lifecycle }}
          lifecycle:
{{ toYaml $values.envoyproxy.lifecycle | indent 12 -}}
          {{- else if $values.containerSpec.lifecycle.enabled }}
          lifecycle:
           {{- if $values.containerSpec.lifecycle.preStop }}
           preStop:
{{ toYaml $values.containerSpec.lifecycle.preStop | indent 12 -}}
           {{- end }}
          {{- end }}
          resources:
{{ toYaml $values.envoyproxy.resources | trim | indent 12 }}
          ports:
            - containerPort: 9901
              protocol: TCP
              name: envoy-admin
              {{- range $index, $element := $values.ContainerPort }}
            - name: envoy-{{ $element.name}}
              containerPort: {{ $element.envoyPort | default (add 8790 $index) }}
              protocol: TCP
              {{- end }}
          command: ["/usr/local/bin/envoy"]
          args: ["-c", "/etc/envoy-config/envoy-config.json", "-l", "info", "--log-format", "[METADATA][%Y-%m-%d %T.%e][%t][%l][%n] %v"]
          volumeMounts:
            - name: {{ $values.envoyproxy.configMapName | default "envoy-config-volume" }}
              mountPath: /etc/envoy-config/
{{- if $values.envoyproxy.readinessProbe}}
          readinessProbe:
{{ toYaml $values.envoyproxy.readinessProbe | indent 12}} 
{{- end }}  
{{- if $values.envoyproxy.livenessProbe}}
          livenessProbe:
{{ toYaml $values.envoyproxy.livenessProbe | indent 12}} 
{{- end }}                     
{{- end}}
{{- if not (empty $values.containers) }}
{{- range $i, $c := $values.containers }}
{{- if .reuseContainerImage}}
        - name: {{ $root.Chart.Name }}-sidecontainer-{{ add1 $i }}
          image: "{{ $values.server.deployment.image }}:{{ $values.server.deployment.image_tag }}"
          imagePullPolicy: {{ $values.image.pullPolicy }}
{{- if not (empty .env) }}
          env:
{{ toYaml .env | indent 12 }}
{{- end }}
 {{- if not (empty .envFrom) }}
          envFrom:
{{ toYaml .env | indent 12 }}
{{- end }}                   
{{- if not (empty .securityContext) }}
          securityContext:
{{ toYaml .securityContext | indent 12 }}
{{- end }}
{{- if not (empty .command) }}
          command:
{{ toYaml .command | indent 12 -}}
{{- end}}
{{- if not (empty .resizePolicy) }}
          resizePolicy:
{{ toYaml .resziePolicy | indent 12}}
{{- end }}
{{- if not (empty .resources) }}
          resources:
{{ toYaml .resources | indent 12 -}}
{{- end}}
{{- if not (empty .volumeMounts) }}
          volumeMounts:
{{ toYaml .volumeMounts | indent 12 -}}
{{- end}}
{{- else}}
        -
{{- $cCopy := deepCopy . }}
{{- if hasKey $cCopy "ports" }}
{{- $newPorts := list }}
{{- range $port := $cCopy.ports }}
  {{- $newPorts = append $newPorts (dict
      "name" (get $port "name")
      "containerPort" (get $port "containerPort")
      "protocol" (get $port "protocol")
  ) }}
{{- end }}
{{- $_ := set $cCopy "ports" $newPorts }}
{{- end }}
{{ toYaml $cCopy | indent 10 }}
{{- end}}
{{- end}}
{{- end}}


      volumes:
  {{- if $values.appMetrics }}
        - name: envoy-config-volume
          configMap:
            name: sidecar-config-{{ template ".Chart.Name .name" $root }}
  {{- end }}
{{- if not (empty $values.volumes) }}
{{ toYaml $values.volumes | trim | indent 8 }}
{{- end }}
{{- if $values.persistentVolumeClaim.name }}
        - name: {{ $values.persistentVolumeClaim.name }}-vol
          persistentVolumeClaim:
            claimName: {{ $values.persistentVolumeClaim.name }}
{{- end}}
      {{- if $values.ConfigMaps.enabled }}
      {{- range $values.ConfigMaps.maps | default (list) }}
      {{- if eq .type "volume"}}
        - name: {{ .name | replace "." "-"}}-vol
          configMap:
            {{- if eq .external true }}
            name: {{ .name }}
            {{- else if eq .external false }}
            name: {{ .name}}-{{ $root.Values.app }}
            {{- end }}
            {{- if eq (len .filePermission) 0 }}
            {{- else }}
            defaultMode: {{ .filePermission}}
            {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}

      {{- if $values.ConfigSecrets.enabled }}
      {{- range $values.ConfigSecrets.secrets | default (list) }}
      {{- if eq .type "volume"}}
        - name: {{ .name | replace "." "-"}}-vol
          secret:
            {{- if eq .external true }}
            secretName: {{ .name }}
            {{- else if eq .external false }}
            secretName: {{ .name}}-{{ $root.Values.app }}
            {{- end }}
            {{- if eq (len .filePermission) 0 }}
            {{- else }}
            defaultMode: {{ .filePermission}}
            {{- end }}
      {{- end }}
      {{- end }}
      {{- end }}
      {{- if and (empty ($values.volumes | default (list))) (eq ($hasPVCExists) false) (or (eq ($values.ConfigSecrets.enabled) true) (eq ($values.ConfigMaps.enabled) true)) (eq ($hasCMVolumeExists) false) (eq ($hasSecretVolumeExists) false) (eq ($values.appMetrics) false) }} []{{- end }}
      {{- if and (empty ($values.volumeMounts | default (list))) (eq ($hasPVCExists) false) (eq ($values.ConfigSecrets.enabled) false) (eq ($values.ConfigMaps.enabled) false) }} [] {{- end }}


  revisionHistoryLimit: 3
##  pauseForSecondsBeforeSwitchActive: {{ $values.pauseForSecondsBeforeSwitchActive }}
#  waitForSecondsBeforeScalingDown: {{ $values.waitForSecondsBeforeScalingDown }}
  strategy:
    {{- if eq $values.deploymentType "ROLLING" }}
    type: "RollingUpdate"
    rollingUpdate:
      maxSurge: {{ $values.deployment.strategy.rolling.maxSurge }}
      maxUnavailable: {{ $values.deployment.strategy.rolling.maxUnavailable }}
    {{- end }}
    {{- if eq $values.deploymentType "RECREATE" }}
    type: "Recreate"
    {{- end }}
{{- end -}}
