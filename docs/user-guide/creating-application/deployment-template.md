
# Deployment Template

Deployment configuration is the Manifest for the application, it defines the runtime behavior of the application. You can define application behavior by providing information in three sections:

1. Chart Version
2. Yaml file
3. Show application metrics

![](../../.gitbook/assets/deployment-template%20%282%29.gif)

### 1. Chart version

| Key | Descriptions |
| :--- | :--- |
| `Chart Version` | Select the Chart Version using which you want to deploy the application. |

Devtron uses helm charts for the deployments. And we are having multiple chart versions based on features it is supporting with every chart version.

One can see multiple chart version options available in the drop-down. you can select any chart version as per your requirements. By default, the latest version of the helm chart is selected in the chart version option.

Every chart version has its own YAML file. Helm charts are used to provide specifications for your application. To make it easy to use, we have created templates for the YAML file and have added some variables inside the YAML. You can provide or change the values of these variables as per your requirement.

If you want to see [Application Metrics](deployment-template.md#3-show-application-metrics) \(For example Status codes 2xx, 3xx, 5xx; throughput, and latency\) for your application, then you need to select the latest chart version.

Application Metrics is not supported for Chart version older than 3.7 version.

### 2. Yaml file

### Container Ports

This defines ports on which application services will be exposed to other services

```yaml
ContainerPort:
    envoyPort: 8799
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
To set environment variables for the containers that run in the Pod.

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
  httpHeader:
  scheme: ""
  tcp: true
```
 
| Key | Description |
| :--- | :--- |
| `Path` | It define the path where the liveness needs to be checked. |
| `failureThreshold` | It defines the maximum number of failures that are acceptable before a given container is not considered as live. |
| `initialDelaySeconds` | It defines the time to wait before a given container is checked for liveliness. |
| `periodSeconds` | It defines the time to check a given container for liveness. |
| `successThreshold` | It defines the number of successes required before a given container is said to fulfil the liveness probe. |
| `timeoutSeconds` | It defines the time for checking timeout. |
| `httpHeader` | Custom headers to set in the request. HTTP allows repeated headers,You can override the default headers by defining .httpHeaders for the probe. |
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
  httpHeader:
  scheme: ""
  tcp: true
```

| Key | Description |
| :--- | :--- |
| `Path` | It define the path where the rediness needs to be checked. |
| `failureThreshold` | It defines the maximum number of failures that are acceptable before a given container is not considered as ready. |
| `initialDelaySeconds` | It defines the time to wait before a given container is checked for readiness. |
| `periodSeconds` | It defines the time to check a given container for readiness. |
| `successThreshold` | It defines the number of successes required before a given container is said to fulfil the rediness probe. |
| `timeoutSeconds` | It defines the time for checking timeout. |
| `httpHeader` | Custom headers to set in the request. HTTP allows repeated headers,You can override the default headers by defining .httpHeaders for the probe. |
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
| `MaxReplicas` | Maximum number of replicas allowed for scaling. |
| `MinReplicas` | Minimum number of replicas allowed for scaling. |
| `TargetCPUUtilizationPercentage` | The target CPU utilization that is expected for a container. |
| `TargetMemoryUtilizationPercentage` | The target memory utilization that is expected for a container. |
| `enabled` | Set true to enable autoscaling else set false.|
| `extraMetrics` | Used to give external metrics for autoscaling. |

### Image

```yaml
image:
  pullPolicy: IfNotPresent
```

Image is used to access images in kubernetes, pullpolicy is used to define the instances calling the image, here the image is pulled when the image is not present,it can also be set as "Always".

### Ingress

This allows public access to the url, please ensure you are using right nginx annotation for nginx class, its default value is nginx

```yaml
ingress:
  enabled: false
  annotations: {}
  path: ""
  host: ""
  tls: []
```

| Key | Description |
| :--- | :--- |
| `enabled` | Enable or disable ingress |
| `annotations` | To configure some options depending on the Ingress controller |
| `path` | Path name |
| `host` | Host name |
| `tls` | It contains security details |

### Ingress Internal

This allows private access to the url, please ensure you are using right nginx annotation for nginx class, its default value is nginx

```yaml
ingressInternal:
  enabled: false
  annotations: {}
  path: ""
  host: ""
  tls: []
```

| Key | Description |
| :--- | :--- |
| `enabled` | Enable or disable ingress |
| `annotations` | To configure some options depending on the Ingress controller |
| `path` | Path name |
| `host` | Host name |
| `tls` | It contains security details |

### Init Containers
```yaml
initContainers: 
  - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        command: ["/usr/local/bin/nginx"]
        args: ["-g", "daemon off;"]
```
Specialized containers that run before app containers in a Pod. Init containers can contain utilities or setup scripts not present in an app image.

### Pause For Seconds Before Switch Active
```yaml
pauseForSecondsBeforeSwitchActive: 30
```
To wait for given period of time before swith active the container.

### Resources

These define minimum and maximum RAM and CPU available to the application.

```yaml
resources:
  limits:
    cpu: '1'
    memory: 200Mi
  requests:
    cpu: '0.10'
    memory: 100Mi
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

### Volumes

```yaml
 volumes: []
```

It is required when some values need to be read from or written to an external disk.

### Volume Mounts

```yaml
volumeMounts: []
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
  key: "key"
  operator: "Equal"
  value: "value"
  effect: "NoSchedule|PreferNoSchedule|NoExecute(1.6 only)"
```

Taints are the opposite, they allow a node to repel a set of pods.

A given pod can access the given node and avoid the given taint only if the given pod satisfies a given taint.

Taints and tolerations work together to ensure that pods are not scheduled onto the inappropriate nodes. One or more taints can be applied to a node, this marks that the node should not accept any pods that don't tolerate the taints.

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
```

It contains the commands for the server.

| Key | Description |
| :--- | :--- |
| `enabled` | To enable or disable the command. |
| `value` | It contains the commands. |


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


## Add on features in Deployment Chart version 3.9.0

### Service Account

```yaml
serviceAccountName: orchestrator
```

A service account provides an identity for the processes that run in a Pod.

When you access the cluster, you are authenticated by the apiserver as a particular User Account. Processes in containers inside pod can also contact the apiserver. When you are authenticated as a particular Service Account.

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

With `maxAvailable` of 1, evictions are allowed as long as atmost 1 unhealthy replica among the total number of desired replicas.

### Application metrics Envoy Configurations

```yaml
envoyproxy:
  image: envoyproxy/envoy:v1.14.1
  configMapName: ""
  resources:
    limits:
      cpu: 50m
      memory: 50Mi
    requests:
      cpu: 50m
      memory: 50Mi
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
```

HPA, by default is configured to work with CPU and Memory metrics. These metrics are useful for internal cluster sizing, but you might want to configure wider set of metrics like service latency, I/O load etc. The custom metrics in HPA can help you to achieve this.

### Wait For Seconds Before Scaling Down
```yaml
waitForSecondsBeforeScalingDown: 30
```
Wait for given period of time before scaling down the container.

### 3. Show application metrics

If you want to see application metrics like different HTTP status codes metrics, application throughput, latency, response time. Enable the Application metrics from below the deployment template Save button. After enabling it, you should be able to see all metrics on App detail page. By default it remains disabled.

![](../../.gitbook/assets/deployment_application_metrics%20%282%29.png)

Once all the Deployment template configurations are done, click on `Save` to save your deployment configuration. Now you are ready to create [Workflow](workflow/) to do CI/CD.

