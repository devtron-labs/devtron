
# Rollout Deployment

Deployment configuration is the Manifest for the application, it defines the runtime behavior of the application. You can define application behavior by providing information in three sections:

1. Chart Version
2. Yaml file
3. Show application metrics

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/deployment-template.gif)

## 1. Chart version

| Key | Descriptions |
| :--- | :--- |
| `Chart Version` | Select the Chart Version using which you want to deploy the application. |

Devtron uses helm charts for the deployments. And we are having multiple chart versions based on features it is supporting with every chart version.

One can see multiple chart version options available in the drop-down. you can select any chart version as per your requirements. By default, the latest version of the helm chart is selected in the chart version option.

Every chart version has its own YAML file. Helm charts are used to provide specifications for your application. To make it easy to use, we have created templates for the YAML file and have added some variables inside the YAML. You can provide or change the values of these variables as per your requirement.

If you want to see [Application Metrics](deployment-template.md#3.-Show-application-metrics) \(For example Status codes 2xx, 3xx, 5xx; throughput, and latency\) for your application, then you need to select the latest chart version.

Application Metrics is not supported for Chart version older than 3.7 version.

## 2. Yaml file

### Container Ports

This defines ports on which application services will be exposed to other services

```yaml
ContainerPort:
  - envoyPort: 8799
    envoyTimeout: 15s
    idleTimeout:
    name: app
    port: 8080
    servicePort: 80
    supportStreaming: true
    useHTTP2: true
```

| Key | Description |
| :--- | :--- |
| `envoyPort` | envoy port for the container. |
| `envoyTimeout` | envoy Timeout for the container,envoy supports a wide range of timeouts that may need to be configured depending on the deployment.By default the envoytimeout is 15s. |
| `idleTimeout` | the duration of time that a connection is idle before the connection is terminated. |
| `name` | name of the port. |
| `port` | port for the container. |
| `servicePort` | port of the corresponding kubernetes service. |
| `supportStreaming` | Used for high performance protocols like grpc where timeout needs to be disabled. |
| `useHTTP2` | Envoy container can accept HTTP2 requests. |

### EnvVariables
```yaml
EnvVariables: []
```
`EnvVariables` provide run-time information to containers and allow to customize how the application works and the behavior of the applications on the system.

Here we can pass the list of env variables , every record is an object which contain the `name` of variable along with `value`.

To set environment variables for the containers that run in the Pod.

### Example of env variables

`IMP` Docker image should have env variables, whatever we want to set.
```yaml
EnvVariables: 
  - name: HOSTNAME
    value: www.xyz.com
  - name: DB_NAME
    value: mydb
  - name: USER_NAME
    value: xyz
```

But `ConfigMap` and `Secret` are the prefered way to inject env variables. So we can create this in `App Configuration` Section

### ConfigMap

It is a centralized storage, specific to k8s namespace where key-value pairs are stored in plain text.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/config-maps/configure-configmap.jpg)

### Secret

It is a centralized storage, specific to k8s namespace where we can store the key-value pairs in plain text as well as in encrypted(`Base64`) form.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/created-secret.gif)

`IMP` All key-values of `Secret` and `CofigMap` will reflect to your application.

### Liveness Probe

If this check fails, kubernetes restarts the pod. This should return error code in case of non-recoverable error.

```yaml
LivenessProbe:
  Path: ""
  port: 8080
  initialDelaySeconds: 20
  periodSeconds: 10
  successThreshold: 1
  timeoutSeconds: 5
  failureThreshold: 3
  command:
    - python
    - /etc/app/healthcheck.py
  httpHeaders:
    - name: Custom-Header
      value: abc
  scheme: ""
  tcp: true
```

| Key | Description |
| :--- | :--- |
| `Path` | It define the path where the liveness needs to be checked. |
| `initialDelaySeconds` | It defines the time to wait before a given container is checked for liveliness. |
| `periodSeconds` | It defines the time to check a given container for liveness. |
| `successThreshold` | It defines the number of successes required before a given container is said to fulfil the liveness probe. |
| `timeoutSeconds` | It defines the time for checking timeout. |
| `failureThreshold` | It defines the maximum number of failures that are acceptable before a given container is not considered as live. |
| `command` | The mentioned command is executed to perform the livenessProbe. If the command returns a non-zero value, it's equivalent to a failed probe. |
| `httpHeaders` | Custom headers to set in the request. HTTP allows repeated headers,You can override the default headers by defining .httpHeaders for the probe. |
| `scheme` | Scheme to use for connecting to the host (HTTP or HTTPS). Defaults to HTTP.
| `tcp` | The kubelet will attempt to open a socket to your container on the specified port. If it can establish a connection, the container is considered healthy. |


### MaxUnavailable

```yaml
  MaxUnavailable: 0
```
The maximum number of pods that can be unavailable during the update process. The value of "MaxUnavailable: " can be an absolute number or percentage of the replicas count. The default value of "MaxUnavailable: " is 25%.

### MaxSurge

```yaml
MaxSurge: 1
```
The maximum number of pods that can be created over the desired number of pods. For "MaxSurge: " also, the value can be an absolute number or percentage of the replicas count.
The default value of "MaxSurge: " is 25%.

### Min Ready Seconds

```yaml
MinReadySeconds: 60
```
This specifies the minimum number of seconds for which a newly created Pod should be ready without any of its containers crashing, for it to be considered available. This defaults to 0 (the Pod will be considered available as soon as it is ready).

### Readiness Probe

If this check fails, kubernetes stops sending traffic to the application. This should return error code in case of errors which can be recovered from if traffic is stopped.

```yaml
ReadinessProbe:
  Path: ""
  port: 8080
  initialDelaySeconds: 20
  periodSeconds: 10
  successThreshold: 1
  timeoutSeconds: 5
  failureThreshold: 3
  command:
    - python
    - /etc/app/healthcheck.py
  httpHeaders:
    - name: Custom-Header
      value: abc
  scheme: ""
  tcp: true
```

| Key | Description |
| :--- | :--- |
| `Path` | It define the path where the readiness needs to be checked. |
| `initialDelaySeconds` | It defines the time to wait before a given container is checked for readiness. |
| `periodSeconds` | It defines the time to check a given container for readiness. |
| `successThreshold` | It defines the number of successes required before a given container is said to fulfill the readiness probe. |
| `timeoutSeconds` | It defines the time for checking timeout. |
| `failureThreshold` | It defines the maximum number of failures that are acceptable before a given container is not considered as ready. |
| `command` | The mentioned command is executed to perform the readinessProbe. If the command returns a non-zero value, it's equivalent to a failed probe. |
| `httpHeaders` | Custom headers to set in the request. HTTP allows repeated headers,You can override the default headers by defining .httpHeaders for the probe. |
| `scheme` | Scheme to use for connecting to the host (HTTP or HTTPS). Defaults to HTTP.
| `tcp` | The kubelet will attempt to open a socket to your container on the specified port. If it can establish a connection, the container is considered healthy. |

### Autoscaling

This is connected to HPA and controls scaling up and down in response to request load.

```yaml
autoscaling:
  enabled: false
  MinReplicas: 1
  MaxReplicas: 2
  TargetCPUUtilizationPercentage: 90
  TargetMemoryUtilizationPercentage: 80
  extraMetrics: []
```

| Key | Description |
| :--- | :--- |
| `enabled` | Set true to enable autoscaling else set false.|
| `MinReplicas` | Minimum number of replicas allowed for scaling. |
| `MaxReplicas` | Maximum number of replicas allowed for scaling. |
| `TargetCPUUtilizationPercentage` | The target CPU utilization that is expected for a container. |
| `TargetMemoryUtilizationPercentage` | The target memory utilization that is expected for a container. |
| `extraMetrics` | Used to give external metrics for autoscaling. |

### Fullname Override

```yaml
fullnameOverride: app-name
```
`fullnameOverride` replaces the release fullname created by default by devtron, which is used to construct Kubernetes object names. By default, devtron uses {app-name}-{environment-name} as release fullname.

### Image

```yaml
image:
  pullPolicy: IfNotPresent
```

Image is used to access images in kubernetes, pullpolicy is used to define the instances calling the image, here the image is pulled when the image is not present,it can also be set as "Always".

### imagePullSecrets

`imagePullSecrets` contains the docker credentials that are used for accessing a registry. 

```yaml
imagePullSecrets:
  - regcred
```
regcred is the secret that contains the docker credentials that are used for accessing a registry. Devtron will not create this secret automatically, you'll have to create this secret using dt-secrets helm chart in the App store or create one using kubectl. You can follow this documentation Pull an Image from a Private Registry [https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/) .

### Ingress

This allows public access to the url, please ensure you are using right nginx annotation for nginx class, its default value is nginx

```yaml
ingress:
  enabled: false
  # For K8s 1.19 and above use ingressClassName instead of annotation kubernetes.io/ingress.class:
  className: nginx
  annotations: {}
  hosts:
      - host: example1.com
        pathType: "ImplementationSpecific"
        paths:
            - /example
      - host: example2.com
        pathType: "ImplementationSpecific"
        paths:
            - /example2
            - /example2/healthz
  tls: []
```
Legacy deployment-template ingress format

```yaml
ingress:
  enabled: false
  # For K8s 1.19 and above use ingressClassName instead of annotation kubernetes.io/ingress.class:
  ingressClassName: nginx-internal
  annotations: {}
  path: ""
  host: ""
  tls: []
```

| Key | Description |
| :--- | :--- |
| `enabled` | Enable or disable ingress |
| `annotations` | To configure some options depending on the Ingress controller |
| `host` | Host name |
| `pathType` | Path in an Ingress is required to have a corresponding path type. Supported path types are `ImplementationSpecific`, `Exact` and `Prefix`. |
| `path` | Path name |
| `tls` | It contains security details |

### Ingress Internal

This allows private access to the url, please ensure you are using right nginx annotation for nginx class, its default value is nginx

```yaml
ingressInternal:
  enabled: false
  # For K8s 1.19 and above use ingressClassName instead of annotation kubernetes.io/ingress.class:
  ingressClassName: nginx-internal
  annotations: {}
  hosts:
      - host: example1.com
        pathType: "ImplementationSpecific"
        paths:
            - /example
      - host: example2.com
        pathType: "ImplementationSpecific"
        paths:
            - /example2
            - /example2/healthz
  tls: []
```

| Key | Description |
| :--- | :--- |
| `enabled` | Enable or disable ingress |
| `annotations` | To configure some options depending on the Ingress controller |
| `host` | Host name |
| `pathType` | Path in an Ingress is required to have a corresponding path type. Supported path types are `ImplementationSpecific`, `Exact` and `Prefix`. |
| `path` | Path name |
| `pathType` | Supported path types are `ImplementationSpecific`, `Exact` and `Prefix`.|
| `tls` | It contains security details |

### Init Containers
```yaml
initContainers: 
  - reuseContainerImage: true
    volumeMounts:
     - mountPath: /etc/ls-oms
       name: ls-oms-cm-vol
   command:
     - flyway
     - -configFiles=/etc/ls-oms/flyway.conf
     - migrate

  - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        command: ["/usr/local/bin/nginx"]
        args: ["-g", "daemon off;"]
```
Specialized containers that run before app containers in a Pod. Init containers can contain utilities or setup scripts not present in an app image. One can use base image inside initContainer by setting the reuseContainerImage flag to `true`.

### Pause For Seconds Before Switch Active
```yaml
pauseForSecondsBeforeSwitchActive: 30
```
To wait for given period of time before switch active the container.

### Resources

These define minimum and maximum RAM and CPU available to the application.

```yaml
resources:
  limits:
    cpu: "1"
    memory: "200Mi"
  requests:
    cpu: "0.10"
    memory: "100Mi"
```

Resources are required to set CPU and memory usage.

#### Limits

Limits make sure a container never goes above a certain value. The container is only allowed to go up to the limit, and then it is restricted.

#### Requests

Requests are what the container is guaranteed to get.

### Service

This defines annotations and the type of service, optionally can define name also.

```yaml
  service:
    type: ClusterIP
    annotations: {}
```
| Key | Description |
| :--- | :--- |
| `type` | Select the type of service, default `ClusterIP` |
| `annotations` | Annotations are widely used to attach metadata and configs in Kubernetes. |
| `name` | Optional field to assign name to service  |
| `loadBalancerSourceRanges` | If service type is `LoadBalancer`, Provide a list of whitelisted IPs CIDR that will be allowed to use the Load Balancer. |

Note - If `loadBalancerSourceRanges` is not set, Kubernetes allows traffic from 0.0.0.0/0 to the LoadBalancer / Node Security Group(s). 


### Volumes

```yaml
volumes:
  - name: log-volume
    emptyDir: {}
  - name: logpv
    persistentVolumeClaim:
      claimName: logpvc
```

It is required when some values need to be read from or written to an external disk.

### Volume Mounts

```yaml
volumeMounts:
  - mountPath: /var/log/nginx/
    name: log-volume 
  - mountPath: /mnt/logs
    name: logpvc
    subPath: employee  
```

It is used to provide mounts to the volume.

### Affinity and anti-affinity

```yaml
Spec:
  Affinity:
    Key:
    Values:
```

Spec is used to define the desire state of the given container.

Node Affinity allows you to constrain which nodes your pod is eligible to schedule on, based on labels of the node.

Inter-pod affinity allow you to constrain which nodes your pod is eligible to be scheduled based on labels on pods.

#### Key

Key part of the label for node selection, this should be same as that on node. Please confirm with devops team.

#### Values

Value part of the label for node selection, this should be same as that on node. Please confirm with devops team.

### Tolerations

```yaml
tolerations:
 - key: "key"
   operator: "Equal"
   value: "value"
   effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
```

Taints are the opposite, they allow a node to repel a set of pods.

A given pod can access the given node and avoid the given taint only if the given pod satisfies a given taint.

Taints and tolerations are a mechanism which work together that allows you to ensure that pods are not placed on inappropriate nodes. Taints are added to nodes, while tolerations are defined in the pod specification. When you taint a node, it will repel all the pods except those that have a toleration for that taint. A node can have one or many taints associated with it.

### Arguments

```yaml
args:
  enabled: false
  value: []
```

This is used to give arguments to command.

### Command

```yaml
command:
  enabled: false
  value: []
  workingDir: {}
```

It contains the commands to run inside the container.

| Key | Description |
| :--- | :--- |
| `enabled` | To enable or disable the command. |
| `value` | It contains the commands. |
| `workingDir` | It is used to specify the working directory where commands will be executed. |

### Containers
Containers section can be used to run side-car containers along with your main container within same pod. Containers running within same pod can share volumes and IP Address and can address each other @localhost.

```yaml
    containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        command: ["/usr/local/bin/nginx"]
        args: ["-g", "daemon off;"]
```

### Prometheus

```yaml
  prometheus:
    release: monitoring
```

It is a kubernetes monitoring tool and the name of the file to be monitored as monitoring in the given case.It describes the state of the prometheus.

### rawYaml

```yaml
rawYaml: 
  - apiVersion: v1
    kind: Service
    metadata:
      name: my-service
    spec:
      selector:
        app: MyApp
      ports:
        - protocol: TCP
          port: 80
          targetPort: 9376
      type: ClusterIP
```
Accepts an array of Kubernetes objects. You can specify any kubernetes yaml here and it will be applied when your app gets deployed.

### Grace Period

```yaml
GracePeriod: 30
```
Kubernetes waits for the specified time called the termination grace period before terminating the pods. By default, this is 30 seconds. If your pod usually takes longer than 30 seconds to shut down gracefully, make sure you increase the `GracePeriod`.

A Graceful termination in practice means that your application needs to handle the SIGTERM message and begin shutting down when it receives it. This means saving all data that needs to be saved, closing down network connections, finishing any work that is left, and other similar tasks.

There are many reasons why Kubernetes might terminate a perfectly healthy container. If you update your deployment with a rolling update, Kubernetes slowly terminates old pods while spinning up new ones. If you drain a node, Kubernetes terminates all pods on that node. If a node runs out of resources, Kubernetes terminates pods to free those resources. It’s important that your application handle termination gracefully so that there is minimal impact on the end user and the time-to-recovery is as fast as possible.


### Server

```yaml
server:
  deployment:
    image_tag: 1-95a53
    image: ""
```

It is used for providing server configurations.

#### Deployment

It gives the details for deployment.

| Key | Description |
| :--- | :--- |
| `image_tag` | It is the image tag |
| `image` | It is the URL of the image |

### Service Monitor

```yaml
servicemonitor:
      enabled: true
      path: /abc
      scheme: 'http'
      interval: 30s
      scrapeTimeout: 20s
      metricRelabelings:
        - sourceLabels: [namespace]
          regex: '(.*)'
          replacement: myapp
          targetLabel: target_namespace
```

It gives the set of targets to be monitored.

### Db Migration Config

```yaml
dbMigrationConfig:
  enabled: false
```

It is used to configure database migration.

### Application Metrics

Application metrics can be enabled to see your application's metrics-CPUService Monito usage,Memory Usage,Status,Throughput and Latency.

### Deployment Metrics

It gives the realtime metrics of the deployed applications

| Key | Description |
| :--- | :--- |
| `Deployment Frequency` | It shows how often this app is deployed to production |
| `Change Failure Rate` | It shows how often the respective pipeline fails. |
| `Mean Lead Time` | It shows the average time taken to deliver a change to production. |
| `Mean Time to Recovery` | It shows the average time taken to fix a failed pipeline. |


## Addon features in Deployment Template Chart version 3.9.0

### Service Account

```yaml
serviceAccountName: orchestrator
```

A service account provides an identity for the processes that run in a Pod.

When you access the cluster, you are authenticated by the API server as a particular User Account. Processes in containers inside pod can also contact the API server. When you are authenticated as a particular Service Account.

When you create a pod, if you do not create a service account, it is automatically assigned the default service account in the namespace.

### Pod Disruption Budget

```yaml
podDisruptionBudget: {}
     minAvailable: 1
     maxUnavailable: 1
```

You can create `PodDisruptionBudget` for each application. A PDB limits the number of pods of a replicated application that are down simultaneously from voluntary disruptions. For example, an application would like to ensure the number of replicas running is never brought below the certain number.

You can specify `maxUnavailable` and `minAvailable` in a `PodDisruptionBudget`.

With `minAvailable` of 1, evictions are allowed as long as they leave behind 1 or more healthy pods of the total number of desired replicas.

With `maxAvailable` of 1, evictions are allowed as long as at most 1 unhealthy replica among the total number of desired replicas.

### Application metrics Envoy Configurations

```yaml
envoyproxy:
  image: envoyproxy/envoy:v1.14.1
  configMapName: ""
  resources:
    limits:
      cpu: "50m"
      memory: "50Mi"
    requests:
      cpu: "50m"
      memory: "50Mi"
```

Envoy is attached as a sidecar to the application container to collect metrics like 4XX, 5XX, Throughput and latency. You can now configure the envoy settings such as idleTimeout, resources etc.

### Prometheus Rule

```yaml
prometheusRule:
  enabled: true
  additionalLabels: {}
  namespace: ""
  rules:
    - alert: TooMany500s
      expr: 100 * ( sum( nginx_ingress_controller_requests{status=~"5.+"} ) / sum(nginx_ingress_controller_requests) ) > 5
      for: 1m
      labels:
        severity: critical
      annotations:
        description: Too many 5XXs
        summary: More than 5% of the all requests did return 5XX, this require your attention
```

Alerting rules allow you to define alert conditions based on Prometheus expressions and to send notifications about firing alerts to an external service.

In this case, Prometheus will check that the alert continues to be active during each evaluation for 1 minute before firing the alert. Elements that are active, but not firing yet, are in the pending state.

### Pod Labels
Labels are key/value pairs that are attached to pods. Labels are intended to be used to specify identifying attributes of objects that are meaningful and relevant to users, but do not directly imply semantics to the core system. Labels can be used to organize and to select subsets of objects.
```yaml
podLabels:
  severity: critical
```

### Pod Annotations
Pod Annotations are widely used to attach metadata and configs in Kubernetes.

```yaml
podAnnotations:
  fluentbit.io/exclude: "true"
```

### Custom Metrics in HPA

```yaml
autoscaling:
  enabled: true
  MinReplicas: 1
  MaxReplicas: 2
  TargetCPUUtilizationPercentage: 90
  TargetMemoryUtilizationPercentage: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

HPA, by default is configured to work with CPU and Memory metrics. These metrics are useful for internal cluster sizing, but you might want to configure wider set of metrics like service latency, I/O load etc. The custom metrics in HPA can help you to achieve this.

### Wait For Seconds Before Scaling Down
```yaml
waitForSecondsBeforeScalingDown: 30
```
Wait for given period of time before scaling down the container.

## 3. Show Application Metrics

If you want to see application metrics like different HTTP status codes metrics, application throughput, latency, response time. Enable the Application metrics from below the deployment template Save button. After enabling it, you should be able to see all metrics on App detail page. By default it remains disabled.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/application-metrics.jpg)

Once all the Deployment template configurations are done, click on `Save` to save your deployment configuration. Now you are ready to create [Workflow](workflow/) to do CI/CD.

### Helm Chart Json Schema Table

Helm Chart json schema is used to validate the deployment template values.

| Chart Version | Link |
| :--- | :--- |
| `reference-chart_3-12-0` | [Json Schema](../../../scripts/devtron-reference-helm-charts/reference-chart_3-12-0/schema.json) |
| `reference-chart_3-11-0` | [Json Schema](../../../scripts/devtron-reference-helm-charts/reference-chart_3-11-0/schema.json) |
| `reference-chart_3-10-0` | [Json Schema](../../../scripts/devtron-reference-helm-charts/reference-chart_3-10-0/schema.json) |
| `reference-chart_3-9-0` | [Json Schema](../../../scripts/devtron-reference-helm-charts/reference-chart_3-9-0/schema.json) |


### Other Validations in Json Schema

The values of CPU and Memory in limits must be greater than or equal to in requests respectively. Similarly, In case of envoyproxy, the values of limits are greater than or equal to requests as mentioned below.
```
resources.limits.cpu >= resources.requests.cpu
resources.limits.memory >= resources.requests.memory
envoyproxy.resources.limits.cpu >= envoyproxy.resources.requests.cpu
envoyproxy.resources.limits.memory >= envoyproxy.resources.requests.memory
```

## Addon features in Deployment Template Chart version 4.11.0

### KEDA Autoscaling
[KEDA](https://keda.sh) is a Kubernetes-based Event Driven Autoscaler. With KEDA, you can drive the scaling of any container in Kubernetes based on the number of events needing to be processed. KEDA can be installed into any Kubernetes cluster and can work alongside standard Kubernetes components like the Horizontal Pod Autoscaler(HPA).

Example for autosccaling with KEDA using Prometheus metrics is given below:
```yaml
kedaAutoscaling:
  enabled: true
  minReplicaCount: 1
  maxReplicaCount: 2
  idleReplicaCount: 0
  pollingInterval: 30
  advanced:
    restoreToOriginalReplicaCount: true
    horizontalPodAutoscalerConfig:
      behavior:
        scaleDown:
          stabilizationWindowSeconds: 300
          policies:
          - type: Percent
            value: 100
            periodSeconds: 15
  triggers: 
    - type: prometheus
      metadata:
        serverAddress:  http://<prometheus-host>:9090
        metricName: http_request_total
        query: envoy_cluster_upstream_rq{appId="300", cluster_name="300-0", container="envoy",}
        threshold: "50"
  triggerAuthentication:
    enabled: false
    name:
    spec: {}
  authenticationRef: {}
```
Example for autosccaling with KEDA based on kafka is given below :
```yaml
kedaAutoscaling:
  enabled: true
  minReplicaCount: 1
  maxReplicaCount: 2
  idleReplicaCount: 0
  pollingInterval: 30
  advanced: {}
  triggers: 
    - type: kafka
      metadata:
        bootstrapServers: b-2.kafka-msk-dev.example.c2.kafka.ap-southeast-1.amazonaws.com:9092,b-3.kafka-msk-dev.example.c2.kafka.ap-southeast-1.amazonaws.com:9092,b-1.kafka-msk-dev.example.c2.kafka.ap-southeast-1.amazonaws.com:9092
        topic: Orders-Service-ESP.info
        lagThreshold: "100"
        consumerGroup: oders-remove-delivered-packages
        allowIdleConsumers: "true"
  triggerAuthentication:
    enabled: true
    name: keda-trigger-auth-kafka-credential
    spec:
      secretTargetRef:
        - parameter: sasl
          name: keda-kafka-secrets
          key: sasl
        - parameter: username
          name: keda-kafka-secrets
          key: username
  authenticationRef: 
    name: keda-trigger-auth-kafka-credential
```

### Security Context
A security context defines privilege and access control settings for a Pod or Container.  

To add a security context for main container:
```yaml
containerSecurityContext:
  allowPrivilegeEscalation: false
```

To add a security context on pod level:
```yaml
podSecurityContext:
  runAsUser: 1000
  runAsGroup: 3000
  fsGroup: 2000
```

### Topology Spread Constraints
You can use topology spread constraints to control how Pods are spread across your cluster among failure-domains such as regions, zones, nodes, and other user-defined topology domains. This can help to achieve high availability as well as efficient resource utilization.

```yaml
topologySpreadConstraints:
  - maxSkew: 1
    topologyKey: zone
    whenUnsatisfiable: DoNotSchedule
    autoLabelSelector: true
    customLabelSelector: {}
```
