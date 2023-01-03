# Resource Browser

`Resource Browser` lists all of the resources running in all of the clusters in your current project. You can use it to view, inspect, manage, and delete resources in your clusters. You can also create resources from the `Resource Browser`.

Resource Browser are helpful for DevOps workflows, troubleshooting issues, and when working with multiple clusters. Rather than using the command-line to query clusters for information about their resources, you can easily get information about all resources in every cluster quickly and easily using it.

You can list and filter resources by specific resource Kinds. You can also preview `Manifest`, `Events`, `Logs` and access `Terminal` by selecting ellipsis on the specific resource.


### Manifest

The Manifest shows the critical information such as container-image, restartCount, state, phase, podIP, startTime etc. and status of the pods which are deployed.

## Events

Events display you the events that took place during the deployment of an app. These events are available until 15 minutes of deployment of the application.

## Logs

Logs contain the logs of the Pods and Containers deployed which you can use for the process of debugging.

## Create Kubernetes Resource

With `Create` button, you can create Kubernetes object by providing the object specification that describes its desired state as well as some basic information about the object (such as a name). You provide the information to in a .yaml file. `kubectl` converts the information to JSON when making the API request automatically.

An example in .yaml file that shows the required fields and object specifications for a Kubernetes Deployment:

```bash
apiVersion: apps/v1

kind: Deployment

metadata:

  name: nginx-deployment

spec:

  selector:

    matchLabels:

      app: nginx

  replicas: 2 # tells deployment to run pods matching the template

  template:

    metadata:

      labels:

        app: nginx

    spec:

      containers:

      - name: nginx

        image: nginx:1.14.2

        ports:

        - containerPort: 80

```






