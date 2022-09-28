# Application Details

## Access an external link

The users can access the [configured external links](../../user-guide/global-configurations/external-links.md) on the **App Details** page.

1. Select **Applications** from the left navigation pane.
2. After selecting a configured application, select the **App Details** tab.
   
> **Note**: The external link configured on the cluster where your app is located is the only one that is visible.

As shown in the screenshot, the monitoring tool appears on the configured component level:

![External links at apps and pod level](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/link-app-pod-level.png)

![External links at container level](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/link-container-level.png)


3. Click on an external link to access the Monitoring Tool.

The link opens in a new tab with the context you specified as `env` variables in the [Add an external link](./global-configurations/../../global-configurations/external-links.md) section.

## Change Project of your Application

You can change your project of your application by selecting **About app** from your application.

![Project Change](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/project-change.png)

The following fields are provided on the **About** page:
| Fields | Description |
| --- | --- |
| **App name** | Shows the name of the app. |
| **Created on** | Shows the day, date and time the app was created. |
| **Created by** | Shows the name of a user. |
| **Project**    | Select the project you want to change from the drop-down list. |

Click **Save**. The selected projected will be updated in your application.

**Note**: If you change the project:
* The current users will lose the access to the application.
* The users who already have an access to the selected project, will get an access to the application automatically.