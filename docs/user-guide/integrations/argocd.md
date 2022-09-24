# GitOps Argo CD Integration with Devtron
 
Devtron integrates deeply with ArgoCD to implement GitOps for continuous delivery. Argo CD follows the GitOps pattern of using Git repositories as the source of truth for defining the desired application state in target Kubernetes cluster. For more information check [Argo CD documentation](https://argo-cd.readthedocs.io/en/stable/)
 
**Features**

* No GitOps plumbing required
* Seamless integration with Devtron CI pipelines and other Devtron integrations.

## Installation

1. On the **Devtron Stack Manager > Discover** page, select the **GitOps (Argo CD) integration**.
2. On the **Discover integrations/GitOps (Argo CD) page**, select **Install**.
 
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
