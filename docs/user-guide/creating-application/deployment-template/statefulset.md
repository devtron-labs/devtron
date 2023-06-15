# StatefulSet

The StatefulSet chart in Devtron allows you to deploy and manage stateful applications. StatefulSet is a Kubernetes resource that provides guarantees about the ordering and uniqueness of Pods during deployment and scaling. 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/sts-chart.jpg)

It supports only `ONDELETE` and `ROLLINGUPDATE` deployment strategy.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/deployment-template/sts-strategy.jpg)


You can select `SatefulSet` chart when you want to use only basic use cases which contain the following:

* **Managing Stateful Applications:** StatefulSets are ideal for managing stateful applications, such as databases or distributed systems, that require stable network identities and persistent storage for each Pod.

* **Ordered Pod Management:** StatefulSets ensure ordered and predictable management of Pods by providing each Pod with a unique and stable hostname based on a defined naming convention and ordinal index.

* **Updating and Scaling Stateful Applications:** StatefulSets support updating and scaling stateful applications by creating new versions of the StatefulSet and performing rolling updates or scaling operations in a controlled manner, ensuring minimal disruption to the application.

* **Persistent Storage:** StatefulSets have built-in mechanisms for handling persistent volumes, allowing each Pod to have its own unique volume claim and storage. This ensures data persistence even when Pods are rescheduled or restarted.

* **Maintaining Pod Identity:** StatefulSets guarantee consistent identity for each Pod throughout its lifecycle. This stability is maintained even if the Pods are rescheduled, allowing applications to rely on stable network identities.

* **Rollback Capability:** StatefulSets provide the ability to rollback to a previous version in case the current state of the application is unstable or encounters issues, ensuring a known working state for the application.

* **Status Monitoring:** StatefulSets offer status information that can be used to monitor the deployment, including the current version, number of replicas, and the readiness of each Pod. This helps in tracking the health and progress of the StatefulSet deployment.

* **Resource Cleanup:** StatefulSets allow for easy cleanup of older versions by deleting StatefulSets and their associated Pods and persistent volumes that are no longer needed, ensuring efficient resource utilization.


You can define application behavior by providing information in the following sections:

| Key | Descriptions |
| :--- | :--- |
| `Chart version` | Select the Chart Version using which you want to deploy the application.<br> Refer [Chart Version](https://docs.devtron.ai/v/v0.5/usage/applications/creating-application/deployment-template/rollout-deployment#1.-chart-version) section for more detail.</br> |
| `Basic Configuration` | You can select the basic deployment configuration for your application on the **Basic** GUI section instead of configuring the YAML file.<br>Refer [Basic Configuration](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/rollout-deployment#2.-basic-configuration) section for more detail.</br>|
| `Advanced (YAML)` | If you want to do additional configurations, then click **Advanced (YAML)** for modifications.<br>Refer [Advanced (YAML)](https://docs.devtron.ai/usage/applications/creating-application/deployment-template/rollout-deployment#3.-advanced-yaml) section for more detail.</br> |
| `Show application metrics` | You can enable `Show application metrics` to see your application's metrics-CPU Service Monitor usage, Memory Usage, Status, Throughput and Latency.<br>Refer [Application Metrics](https://docs.devtron.ai/v/v0.5/usage/applications/app-details/app-metrics) for more detail.</br> |