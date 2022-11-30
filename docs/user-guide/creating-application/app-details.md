# Application Details

## Access an External Link

The users can access the [configured external links](../../user-guide/global-configurations/external-links.md) on the **App Details** page.

1. Select **Applications** from the left navigation pane.
2. After selecting a configured application, select the **App Details** tab.
   
> **Note**: The external link configured on the cluster where your app is located is the only one that is visible.

As shown in the screenshot, the monitoring tool appears at the configured component level:

![External links at apps and pod level](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/link-app-pod-level.png)

![External links at container level](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/link-container-level.png)


3. Click on an external link to access the Monitoring Tool.

The link opens in a new tab with the context you specified as env variables in the [Add an external link](./global-configurations/../../global-configurations/external-links.md) section.

## Change Project of your Application

You can change the project of your application by selecting **About app** from your application.

![About app](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/about-app3.png)

![Project Change](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/project-change.png)

The following fields are provided on the **About** page:

| Fields | Description |
| :---    |     :---       |
| **App Name**  | Displays the name of the application. |
| **Created on** | Displays the day, date and time the application was created. |
| **Created by**  | Displays the email address of a user who created the application. |
| **Project**   | Displays the currect project of the application. You can change the project by selecting a different project from the drop-down list. |

Click **Save**. The application will be moved to the selected project.

**Note**: If you change the project:
* The current users will lose the access to the application.
* The users who already have an access to the selected project, will get an access to the application automatically.


## Ingress Host URL

You can view the Ingress Host URL and the Load Balancer URL on the **URLs** section on the **App Details**.
You can also copy the Ingress Host URL from the **URLs** instead of searching in the `Manifest`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/ingress-url-appdetails.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/debugging-deployment-and-monitoring/ingress-host-url1.jpg)

1. Select **Applications** from the left navigation pane.
2. After selecting your configured application, select the **App Details**.
3. Click **URLs**.
4. You can view or copy the **URL** of the Ingress Host.


**Note**: 
* The Ingress Host URL will point to the load balancer of your application.
* You can also view the `Service` name with the load balancer detail.


