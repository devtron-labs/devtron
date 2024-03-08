# Monitoring (Grafana)

**Prerequisite**: Please make sure to install **Build and Deploy (CI/CD)** integration. To install it, click [here](../integrations/build-and-deploy-ci-cd.md).
 
Devtron leverages the power of Grafana to show application metrics like CPU, Memory utilization, Status 4xx/ 5xx/ 2xx, Throughput, and Latency. For more information check [Grafana documentation](https://grafana.com/docs/grafana/latest/).
 
**Features**

* CPU usage: Displays the overall utilization of CPU by an application. It is available as aggregated or per pod.
* Memory usage: Displays the overall utilization of memory by an application. It is available as aggregated or per pod.
* Throughput: Indicates the number of requests processed by an application per minute.
* Status codes: Indicates the application’s response to the client’s request with a specific status code as shown below:
       * 1xx: Communicates transfer protocol level information
       * 2xx: Client’s request is processed successfully
       * 3xx: Client must take some additional action to complete their request
       * 4xx: There is an error on the client side
       * 5xx: There is an error on the server side


## Installation

1. On the **Devtron Stack Manager > Discover** page, click the **Monitoring (Grafana)**.
2. On the **Discover integrations/Monitoring (Grafana)** page, click **Install**.
 
The installation status may be one of the following:
 
| Installation status | Description |
| --- | --- |
| Install | The integration is not yet installed. |
| Initializing | The installation is being initialized. |
| Installing | The installation is in progress. The logs are available to track the progress. |
| Failed | Installation failed and the logs are available to troubleshoot. You can retry the installation or [contact support](https://discord.devtron.ai/). |
| Installed | The integration is successfully installed and available on the **Installed** page. |
| Request timed out | The request to install has hit the maximum number of retries. You may retry the installation or [contact support](https://discord.devtron.ai/) for further assistance. |
 
> A list of installed integrations can be viewed on the **Devtron Stack Manager > Installed** page.
 
To update an installed integration, please [update Devtron](../../setup/upgrade/upgrade-devtron-ui.md).
