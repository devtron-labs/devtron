# Using Ephemeral Containers on Devtron

## Introduction

Ephemeral container is a special type of container that runs temporarily in an existing Pod to accomplish user-initiated actions such as troubleshooting. It is especially useful when `kubectl exec` is insufficient because a container has crashed or a container image doesn't include debugging utilities.

For instance, ephemeral containers help you execute a `curl` request from within pods that typically lack this utility.

{% hint style="info" %}
Ephemeral containers are turned on by default in Kubernetes v1.23 and later
{% endhint %}

<hr />

## How to Launch an Ephemeral Container

Wherever you can access pod resources in Devtron, you can launch an ephemeral container as shown below.

### From Devtron (App Details)

{% embed url="https://www.youtube.com/watch?v=TnaHRugYvSI" caption="Launching Ephemeral Container from App Details" %}

1. In the left sidebar, go to **Applications**.
2. Search and click your application from the list of Devtron Apps.
3. Go to the **App Details** tab.
4. Under the **K8 Resources** tab, select **Pod** inside `Workloads`.
5. Locate the pod you wish to debug. Hover and choose click **Terminal**.
    ![Figure 1: Opening a Terminal](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/terminal.jpg)
6. Click **Launch Ephemeral Container** as shown below.
    ![Figure 2: Launching an Ephemeral Container](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/launch-ec-new.jpg)
7. You get 2 tabs:
    * **Basic** - It provides the bare minimum configurations required to launch an ephemeral container.
    ![Figure 3: Basic View](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/basic.jpg)
    It contains 3 mandatory fields:
        * **Container name prefix** - Type a prefix to give to your ephemeral container, for e.g., *debug*. Your container name would look like `debug-jndvs`.
        * **Image** - Choose an image to run from the dropdown. Ephemeral containers need an image to run and provide the capability to debug, such as `curl`. You can use a custom image too.
        * **Target Container name** - Since a pod can have one or more containers, choose a target container you wish to debug, from the dropdown.

    * **Advanced** - It is particularly useful for advanced users that wish to use labels or annotations since it provides additional key-value options. Refer [Ephemeral Container Spec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#ephemeralcontainer-v1-core) to view the supported options.
    ![Figure 4: Advanced View](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/advanced.jpg)
    
    {% hint style="info" %}
    Devtron ignores the `command` field while launching an ephemeral container
    {% endhint %}
8. Click **Launch Container**.

### From Devtron (Resource Browser)

{% embed url="https://www.youtube.com/watch?v=Ml19i29Ivc4" caption="Launching Ephemeral Containers from Resource Browser" %}

1. In the left sidebar, go to **Resource Browser**.
2. Select **Pod** inside `Workloads`.
3. Use the searchbar to find and locate the pod you wish to debug. Click the pod.
4. Go to the **Terminal** tab 

The rest is same as steps 6, 7, and 8 mentioned in the [previous section](#from-devtron-app-details).

### From Devtron's Cluster Terminal

(*This is not a recommended method. This option is available only if you are an admin.*)

{% embed url="https://www.youtube.com/watch?v=PzB6dFRYe38" caption="Externally Created Ephemeral Container" %}

You can launch an ephemeral container from Kubernetes CLI. For this, you need access to the cluster terminal on Devtron.

<hr />

## Removing an Ephemeral Container

{% embed url="https://www.youtube.com/watch?v=tZID0YU0YUU" caption="Deleting Ephemeral Containers" %}

You can remove an ephemeral container using either **App Details** or **Resource Browser** (from the same screen you used to create the ephemeral container).

{% hint style="info" %}
You cannot use **App Details** or **Resource Browser** to remove an ephemeral container created using Kubernetes CLI
{% endhint %}

![Figure 5: Removing Ephemeral Container from App Details](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/delete-ec.jpg)





