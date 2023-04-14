# Application Details

## Access an External Link

The users can access the [configured external links](../../user-guide/global-configurations/external-links.md) on the **App Details** page.

1. Select **Applications** from the left navigation pane.
2. After selecting a configured application, select the **App Details** tab.
   
> **Note**: If you enable `App admins can edit` on the `External Links` page, then only non-super admin users can view the selected links on the `App-Details` page.

As shown in the screenshot, the external links appear on the `App-Details` level:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/external-links/app-details-external-link.png)


3. You can hover around an external link (e.g. Grafana) to view the description.

The link opens in a new tab with the context you specified as env variables in the [Add an external link](./global-configurations/../../global-configurations/external-links.md) section.


## Manage External Links

On the `App Configuration` page, select `External Links` from the navigation pane.
You can see the configured external links which can be searched, edited or deleted.

You can also `Add Link` to add a new external link.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/external-links/app-config-external-link.jpg)



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


