
# StatefulSet Chart 1.0.0

## 1. Yaml File -

### Container Ports

This defines ports on which application services will be exposed to other services

```yaml
ContainerPort:
  - envoyPort: 8799
    idleTimeout:
    name: app
    port: 8080
    servicePort: 80
    nodePort: 32056
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
| `nodePort` | nodeport of the corresponding kubernetes service. |
| `supportStreaming` | Used for high performance protocols like grpc where timeout needs to be disabled. |
| `useHTTP2` | Envoy container can accept HTTP2 requests. |

### EnvVariables
```yaml
EnvVariables: []
```

### EnvVariablesFromSecretKeys
```yaml
EnvVariablesFromSecretKeys: 
  - name: ENV_NAME
    secretName: SECRET_NAME
    keyName: SECRET_KEY

```
 It is use to get the name of Environment Variable name, Secret name and the Key name from which we are using the value in that corresponding Environment Variable.

 ### EnvVariablesFromCongigMapKeys
```yaml
EnvVariablesFromCongigMapKeys: 
  - name: ENV_NAME
    configMapName: CONFIG_MAP_NAME
    keyName: CONFIG_MAP_KEY

```
 It is use to get the name of Environment Variable name, Config Map name and the Key name from which we are using the value in that corresponding Environment Variable.

To set environment variables for the containers that run in the Pod.
### StatefulSetConfig

| Key | Description |
| :--- | :--- |
| `labels` |  set of key-value pairs used to identify the StatefulSet . |
| `annotations` | A map of key-value pairs that are attached to the stateful set as metadata. |
| `serviceName` | The name of the Kubernetes Service that the StatefulSet should create. |
| `podManagementPolicy` | A policy that determines how Pods are created and deleted by the StatefulSet. In this case, the policy is set to "Parallel", which means that all Pods are created at once. |
| `revisionHistoryLimit` | The number of revisions that should be stored for each replica of the StatefulSet. |
| `updateStrategy` | The update strategy used by the StatefulSet when rolling out changes. |
| `mountPath` | The path where the volume should be mounted in the container. |

volumeClaimTemplates: An array of volume claim templates that are used to create persistent volumes for the StatefulSet. Each volume claim template specifies the storage class, access mode, storage size, and other details of the persistent volume.


| Key | Description |
| :--- | :--- |
| `apiVersion` |  The API version of the PVC . |
| `kind` | The type of object that the PVC is. |
| `metadata` | Metadata that is attached to the resource being created. |
| `labels` | A set of key-value pairs used to label the object for identification and selection. |
| `spec` | The specification of the object, which defines its desired state and behavior.|
| `accessModes` | A list of access modes for the PersistentVolumeClaim, such as "ReadWriteOnce" or "ReadWriteMany". |
| `dataSource` | A data source used to populate the PersistentVolumeClaim, such as a Snapshot or a StorageClass. |
| `kind`| specifies the kind of the snapshot, in this case Snapshot.|
| `apiGroup`| specifies the API group of the snapshot API, in this case snapshot.storage.k8s.io.|
| `name`| specifies the name of the snapshot, in this case my-snapshot.|
| `dataSourceRef` | A  reference to a data source used to create the persistent volume. In this case, it's a secret. |
| `updateStrategy` | The update strategy used by the StatefulSet when rolling out changes. |
| `resources` | The resource requests and limits for the PersistentVolumeClaim, which define the minimum and maximum amount of storage it can use. |
| `requests` | The amount of storage requested by the PersistentVolumeClaim. |
| `limits` | The maximum amount of storage that the PersistentVolumeClaim can use. |
| `storageClassName` | The name of the storage class to use for the persistent volume. |
| `selector` | The selector used to match a persistent volume to a persistent volume claim. |
| `matchLabels` | a map of key-value pairs to match the labels of the corresponding PersistentVolume.|
| `matchExpressions` |A set of requirements that the selected object must meet to be considered a match. |
| `key` | The key of the label or annotation to match.|
| `operator` | The operator used to compare the key-value pairs (in this case, "In" specifies a set membership test).|
| `values` | A list of values that the selected object's label or annotation must match.|
| `volumeMode` | The mode of the volume, either "Filesystem" or "Block". |
| `volumeName` | The name of the PersistentVolume that is created for the PersistentVolumeClaim. |
These are  all the configuration settings for the StatefulSet.
```yaml
statefulSetConfig:
  labels:
    app: my-statefulset
    environment: production
  annotations:
    example.com/version: "1.0"
  serviceName: "my-statefulset-service"
  podManagementPolicy: "Parallel"
  revisionHistoryLimit: 5
  mountPath: "/data"
  volumeClaimTemplates:
    - apiVersion: v1
      kind: PersistentVolumeClaim
      metadata:
        labels:
          app: my-statefulset
      spec:
        accessModes:
          - ReadWriteOnce
        dataSource:
          kind: Snapshot
          apiGroup: snapshot.storage.k8s.io
          name: my-snapshot
        resources:
          requests:
            storage: 5Gi
          limits:
            storage: 10Gi
        storageClassName: my-storage-class
        selector:
          matchLabels:
            app: my-statefulset
        volumeMode: Filesystem
        volumeName: my-pv
  - apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: pvc-logs
      labels:
        app: myapp
    spec:
      accessModes:
        - ReadWriteMany
      dataSourceRef:
        kind: Secret
        apiGroup: v1
        name: my-secret
      resources:
        requests:
          storage: 5Gi
      storageClassName: my-storage-class
      selector:
        matchExpressions:
          - {key: environment, operator: In, values: [production]}
      volumeMode: Block
      volumeName: my-pv

```


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
| `httpHeaders` | Custom headers to set in the request. HTTP allows repeated headers,You can override the default headers by defining .httpHeaders for the probe. |
| `scheme` | Scheme to use for connecting to the host (HTTP or HTTPS). Defaults to HTTP.
| `tcp` | The kubelet will attempt to open a socket to your container on the specified port. If it can establish a connection, the container is considered healthy. |

### Ambassador Mappings

You can create ambassador mappings to access your applications from outside the cluster. At its core a Mapping resource maps a resource to a service.

```yaml
ambassadorMapping:
  ambassadorId: "prod-emissary"
  cors: {}
  enabled: true
  hostname: devtron.example.com
  labels: {}
  prefix: /
  retryPolicy: {}
  rewrite: ""
  tls:
    context: "devtron-tls-context"
    create: false
    hosts: []
    secretName: ""
```

| Key | Description |
| :--- | :--- |
| `enabled` | Set true to enable ambassador mapping else set false.|
| `ambassadorId` | used to specify id for specific ambassador mappings controller. |
| `cors` | used to specify cors policy to access host for this mapping. |
| `weight` | used to specify weight for canary ambassador mappings. |
| `hostname` | used to specify hostname for ambassador mapping. |
| `prefix` | used to specify path for ambassador mapping. |
| `labels` | used to provide custom labels for ambassador mapping. |
| `retryPolicy` | used to specify retry policy for ambassador mapping. |
| `corsPolicy` | Provide cors headers on flagger resource. |
| `rewrite` | used to specify whether to redirect the path of this mapping and where. |
| `tls` | used to create or define ambassador TLSContext resource. |
| `extraSpec` | used to provide extra spec values which not present in deployment template for ambassador resource. |

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
        paths:
            - /example
      - host: example2.com
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
| `path` | Path name |
| `host` | Host name |
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
        paths:
            - /example
      - host: example2.com
        paths:
            - /example2
            - /example2/healthz
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
  - reuseContainerImage: true
    securityContext:
      runAsUser: 1000
      runAsGroup: 3000
      fsGroup: 2000
    volumeMounts:
     - mountPath: /etc/ls-oms
       name: ls-oms-cm-vol
   command:
     - flyway
     - -configFiles=/etc/ls-oms/flyway.conf
     - migrate

  - name: nginx
    image: nginx:1.14.2
    securityContext:
      privileged: true
    ports:
    - containerPort: 80
    command: ["/usr/local/bin/nginx"]
    args: ["-g", "daemon off;"]
```
Specialized containers that run before app containers in a Pod. Init containers can contain utilities or setup scripts not present in an app image. One can use base image inside initContainer by setting the reuseContainerImage flag to `true`.

### Istio

Istio is a service mesh which simplifies observability, traffic management, security and much more with it's virtual services and gateways.

```yaml
istio:
  enable: true
  gateway:
    annotations: {}
    enabled: false
    host: example.com
    labels: {}
    tls:
      enabled: false
      secretName: example-tls-secret
  virtualService:
    annotations: {}
    enabled: false
    gateways: []
    hosts: []
    http:
      - corsPolicy:
          allowCredentials: false
          allowHeaders:
            - x-some-header
          allowMethods:
            - GET
          allowOrigin:
            - example.com
          maxAge: 24h
        headers:
          request:
            add:
              x-some-header: value
        match:
          - uri:
              prefix: /v1
          - uri:
              prefix: /v2
        retries:
          attempts: 2
          perTryTimeout: 3s
        rewriteUri: /
        route:
          - destination:
              host: service1
              port: 80
        timeout: 12s
      - route:
          - destination:
              host: service2
    labels: {}
```

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
```

It contains the commands for the server.

| Key | Description |
| :--- | :--- |
| `enabled` | To enable or disable the command. |
| `value` | It contains the commands. |


### Containers
Containers section can be used to run side-car containers along with your main container within same pod. Containers running within same pod can share volumes and IP Address and can address each other @localhost. We can use base image inside container by setting the reuseContainerImage flag to `true`.

```yaml
    containers:
      - name: nginx
        image: nginx:1.14.2
        ports:
        - containerPort: 80
        command: ["/usr/local/bin/nginx"]
        args: ["-g", "daemon off;"]
      - reuseContainerImage: true
        securityContext:
          runAsUser: 1000
          runAsGroup: 3000
          fsGroup: 2000
        volumeMounts:
        - mountPath: /etc/ls-oms
          name: ls-oms-cm-vol
        command:
          - flyway
          - -configFiles=/etc/ls-oms/flyway.conf
          - migrate
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
### Winter-Soldier
Winter Soldier can be used to
- cleans up (delete) Kubernetes resources
- reduce workload pods to 0

**_NOTE:_** After deploying this we can create the Hibernator object and provide the custom configuration by which workloads going to delete, sleep and many more.   for more information check [the main repo](https://github.com/devtron-labs/winter-soldier)

Given below is template values you can give in winter-soldier:
```yaml
winterSoilder:
  enable: false
  apiVersion: pincher.devtron.ai/v1alpha1
  action: sleep
  timeRangesWithZone:
    timeZone: "Asia/Kolkata"
    timeRanges: []
  targetReplicas: []
  fieldSelector: []
```
Here, 
| Key | values | Description |
| :--- | :--- | :--- |
| `enable` | `fasle`,`true` | decide the enabling factor  |
| `apiVersion` | `pincher.devtron.ai/v1beta1`, `pincher.devtron.ai/v1alpha1` | specific api version  |
| `action` | `sleep`,`delete`, `scale` | This specify  the action need to perform.  |
| `timeRangesWithZone`:`timeZone` | eg:- `"Asia/Kolkata"`,`"US/Pacific"` |  It use to specify the timeZone used. (It uses standard format. please refer [this](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones))  |
| `timeRangesWithZone`:`timeRanges` | array of [ `timeFrom`, `timeTo`, `weekdayFrom`, `weekdayTo`] |  It use to define time period/range on which the user need to perform the specified action. you can have multiple timeRanges. <br /> These settings will take `action` on Sat and Sun from 00:00 to 23:59:59, |
| `targetReplicas` | `[n]` : n - number of replicas to scale. | These is mandatory field when the `action` is `scale` <br /> Defalut value is `[]`.  |
| `fieldSelector` | `- AfterTime(AddTime( ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '5m'), Now()) `  | These value will take a list of methods to select the resources on which we perform specified `action` .  |


here is an example,
```yaml
winterSoilder:
  apiVersion: pincher.devtron.ai/v1alpha1 
  enable: true
  annotations: {}
  labels: {}
  timeRangesWithZone:
    timeZone: "Asia/Kolkata"
    timeRanges: 
      - timeFrom: 00:00
        timeTo: 23:59:59
        weekdayFrom: Sat
        weekdayTo: Sun
      - timeFrom: 00:00
        timeTo: 08:00
        weekdayFrom: Mon
        weekdayTo: Fri
      - timeFrom: 20:00
        timeTo: 23:59:59
        weekdayFrom: Mon
        weekdayTo: Fri
  action: scale
  targetReplicas: [1,1,1]
  fieldSelector: 
    - AfterTime(AddTime( ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '10h'), Now())
```
Above settings will take action on `Sat` and `Sun` from 00:00 to 23:59:59, and on `Mon`-`Fri` from 00:00 to 08:00 and 20:00 to 23:59:59. If `action:sleep` then runs hibernate at timeFrom and unhibernate at `timeTo`. If `action: delete` then it will delete workloads at `timeFrom` and `timeTo`. Here the `action:scale` thus it scale the number of resource replicas to  `targetReplicas: [1,1,1]`. Here each element of `targetReplicas` array is mapped with the corresponding elments of array `timeRangesWithZone/timeRanges`. Thus make sure the length of both array is equal, otherwise the cnages cannot be observed.

The above example will select the application objects which have been created 10 hours ago across all namespaces excluding application's namespace. Winter soldier exposes following functions to handle time, cpu and memory.

- ParseTime - This function can be used to parse time. For eg to parse creationTimestamp use ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z')
- AddTime - This can be used to add time. For eg AddTime(ParseTime({{metadata.creationTimestamp}}, '2006-01-02T15:04:05Z'), '-10h') ll add 10h to the time. Use d for day, h for hour, m for minutes and s for seconds. Use negative number to get earlier time.
- Now - This can be used to get current time.
- CpuToNumber - This can be used to compare CPU. For eg any({{spec.containers.#.resources.requests}}, { MemoryToNumber(.memory) < MemoryToNumber('60Mi')}) will check if any resource.requests is less than 60Mi.



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

### Deployment Metrics

It gives the realtime metrics of the deployed applications

| Key | Description |
| :--- | :--- |
| `Deployment Frequency` | It shows how often this app is deployed to production |
| `Change Failure Rate` | It shows how often the respective pipeline fails. |
| `Mean Lead Time` | It shows the average time taken to deliver a change to production. |
| `Mean Time to Recovery` | It shows the average time taken to fix a failed pipeline. |

## 2. Show application metrics

If you want to see application metrics like different HTTP status codes metrics, application throughput, latency, response time. Enable the Application metrics from below the deployment template Save button. After enabling it, you should be able to see all metrics on App detail page. By default it remains disabled.
![](../../../.gitbook/assets/deployment_application_metrics%20%282%29.png)

Once all the Deployment template configurations are done, click on `Save` to save your deployment configuration. Now you are ready to create [Workflow](workflow/) to do CI/CD.

### Helm Chart Json Schema 

Helm Chart [json schema](../../../scripts/devtron-reference-helm-charts/reference-chart_4-11-0/schema.json) is used to validate the deployment template values.

### Other Validations in Json Schema

The values of CPU and Memory in limits must be greater than or equal to in requests respectively. Similarly, In case of envoyproxy, the values of limits are greater than or equal to requests as mentioned below.
```
resources.limits.cpu >= resources.requests.cpu
resources.limits.memory >= resources.requests.memory
envoyproxy.resources.limits.cpu >= envoyproxy.resources.requests.cpu
envoyproxy.resources.limits.memory >= envoyproxy.resources.requests.memory
```
