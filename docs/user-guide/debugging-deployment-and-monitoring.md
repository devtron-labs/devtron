# Debugging Deployment And Monitoring

If the deployment of your application is not successful, then debugging needs to be done to check the cause of the error.

This can be done through `App Details` section which you can access in the following way:-

Applications-&gt;AppName-&gt;App Details

Over here, you can see the status of the app as Healthy. If there are some errors with deployment then the status would not be in a Healthy state.

### Events

![](../.gitbook/assets/events1%20%281%29.jpg)

Events of the application are accessible from the bottom left corner.

Events section displays you the events that took place during the deployment of an app. These events are available until 15 minutes of deployment of the application.

### Logs

![](../.gitbook/assets/events2%20%281%29.jpg)

Logs contain the logs of the Pods and Containers deployed which you can use for the process of debugging.

### Manifest

![](../.gitbook/assets/events3%20%282%29.jpg)

The Manifest shows the critical information such as Container-image, restartCount, state, phase, podIP, startTime etc. and status of the pods deployed.

### Deleting Pods

![](../.gitbook/assets/events5%20%281%29.png)

You might run into a situation where you need to delete Pods. You may need to bounce or restart a pod.

Deleting a Pod is not an irksome task, it can simply be deleted by Clicking on `Delete Pod`.

Suppose you want to setup a new environment, you can delete a pod and thereafter a new pod will be created automatically depending upon the replica count.

### Application Objects

You can view `Application Objects` in this section of `App Details`, such as:

| Key | Description |
| :--- | :--- |
| `Workloads` | _ReplicaSet_\(ensures how many replica of pod should be running\), _Status of Pod_\(status of the Pod\) |
| `Networking` | _Service_\(an abstraction which defines a logical set of Pods\), _Endpoints_\(names of the endpoints that implement a Service\), _Ingress_\(API object that manages external access to the services in a cluster\) |
| `Config & Storage` | _ConfigMap_\( API object used to store non-confidential data in key-value pairs\) |
| `Custom Resource` | _Rollout_\(new Pods will be scheduled on Nodes with available resources\), _ServiceMonitor_\(specifies how groups of services should be monitored\) |

![](../.gitbook/assets/app-details-application-object-ingress.png)

## Monitoring

![](../.gitbook/assets/events4%20%282%29.jpg)

You can monitor the application in the `App Details`section.

Metrics like CPU Usage, Memory Usage, Throughput and Latency can be viewed here.

| Key | Description |
| :--- | :--- |
| `CPU Usage` | Percentage of CPU's cycles used by the app. |
| `Memory Usage` | Amount of memory used by app. |
| `Throughput` | Performance of the app. |
| `Latency` | Delay caused while transmitting the data. |

