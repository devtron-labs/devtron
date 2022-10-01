# Security Integration (Clair)

Devtron security integration **(Clair)** enables you to scan the vulnerabilities of the images at the time of image build only.

While building the container images, it is important to take care of how secure the application is, before it is deployed. In the source code , the third party libraries and outdated libraries might be  used that can add vulnerabilities to the image we deploy. Devtron provides **Clair** integratigration for scanning vulnerabilites of the images.

**Features:**

* You can enable image scan for images if it is required.
* You can set security policies according to your use-cases.
* Blocks deployment for images with vulnerabilities if it is set to `block` from security policies.
* Inherits the security policy (Global / Cluster / Environment / Application) to allow / block vulnerabilities based on criticality (High / Moderate / Low).
* Compares the vulnerabilities against a whitelist.
* Shows security vulnerabilities detected in deployed applications


## Installation

1. On the **Devtron Stack Manager > Discover** page, select the **Vulnerability scanning (Clair) integration**.
2. On the **Discover integrations/Vulnerability scanning (Clair) page**, select **Install**.
 
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