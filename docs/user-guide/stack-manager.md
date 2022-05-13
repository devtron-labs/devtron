# Devtron Integrations
 
Devtron integrations extend the functionality of your Devtron stack.

## Discover and install integrations
 
The current release of Devtron supports the Build and Deploy (CI/CD) integration. More integrations will be available soon; to request one, please [submit a ticket](https://github.com/devtron-labs/devtron/issues/new/choose)

> Integrations can be installed by super admins; However other user roles can browse and request super admins to install the required integrations.

> Integrations are updated along with [Devtron updates](setup/../../setup/upgrade-devtron.md).

Select **Devtron Stack Manager** from the left navigation bar.
Under **INTEGRATIONS**, select **Discover**.
 
![Discover integrations](https://devtron-public-asset.s3.us-east-2.amazonaws.com/integrations/discover-integrations.png)
 
> Although the integrations are installed separately, they cannot be upgraded separately. Integrations update happens automatically with [Devtron upgrade](#upgrade-devtron).
 
### Build and Deploy (CI/CD) integration
 
Devtron CI/CD integration enables software development teams to automate the build and deployment process, allowing them to focus on meeting the business requirements, maintaining code quality, and ensuring security.
 
**Features**
 
* Leverages Kubernetes auto-scaling and centralized caching to give you unlimited cost-efficient CI workers.
* Supports pre-CI and post-CI integrations for code quality monitoring.
* Seamlessly integrates with Clair for image vulnerability scanning.
* Supports different deployment strategies: Blue/Green, Rolling, Canary, and Recreate.
* Implements GitOps to manage the state of Kubernetes applications.
* Integrates with ArgoCD for continuous deployment.
* Check logs, events, and manifests or exec inside containers for debugging.
* Provides deployment metrics like; deployment frequency, lead time, change failure rate, and mean-time recovery.
* Seamless integration with Grafana for continuous application metrics like CPU and memory usage, status code, throughput, and latency on the dashboard.

#### Installation

1. On the **Devtron Stack Manager > Discover** page, select the **Build and Deploy (CI/CD) integration**.
2. On the **Discover integrations/Build and Deploy (CI/CD) page**, select **Install**.
 
The installation status may be one of the following:
 
| Installation status | Description |
| --- | --- |
| Install | The integration is not yet installed. |
| Initializing | The installation is being initialized. |
| Installing | The installation is in progress. The logs are available to track the progress. |
| Failed | Installation failed and the logs are available to troubleshoot. You could retry the installation or [contact support](https://discord.devtron.ai/). |
| Installed | The integration is successfully installed and available on the **Installed page**. |
| Request timed out | The request to install has hit the maximum number of retries. You may retry the installation or [contact support](https://discord.devtron.ai/) for further assistance. |
 
> A list of installed integrations can be viewed on the **Devtron Stack Manager > Installed** page.
 
To update an installed integration, please [update Devtron](../setup/upgrade/upgrade-devtron-ui.md).
