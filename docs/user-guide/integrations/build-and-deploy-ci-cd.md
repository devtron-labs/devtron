# Build and Deploy (CI/CD)
 
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

{% hint style="info" %}
Devtron also gives you the option of partial cloning. It increases the cloning speed of your [code repository](../../../docs/reference/glossary.md#repo), thus reducing the [build time](../../../docs/reference/glossary.md#build-pipeline) during the [CI process](../deploying-application/triggering-ci.md).
{% endhint %}

## Installation

1. On the **Devtron Stack Manager > Discover** page, click the **Build and Deploy (CI/CD)**.
2. On the **Discover Integrations/Build and Deploy (CI/CD)** page, click **Install**.
 
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