# Notifications
 
 **Prerequisite**: Please make sure to install **Build and Deploy (CI/CD)** integration. To install it, click [here](../integrations/build-and-deploy-ci-cd.md).

With Notifications integration, you can receive alerts for build and deployment pipelines on trigger, success, and failure events. An alert will be sent to your desired slack channel and email address (supports SES and SMTP configurations) with the required information to take up the actions, if required.
 
**Features**

* Receive alerts for start, success, and failure events on desired build pipelines.
* Receive alerts for start, success, and failure events on desired deployment pipelines.
* Receive alerts on desired Slack channels via webhook.
* Receive alerts on your email address (supports SES and SMTP).

## Installation

1. On the **Devtron Stack Manager > Discover** page, click the **Notifications**.
2. On the **Discover Integrations/Notifications** page, click **Install**.
 
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
