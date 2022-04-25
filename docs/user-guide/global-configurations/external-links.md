# Link to external Monitoring Tools

External links allow you to connect to the third-party Monitoring Tools within your Devtron dashboard for seamlessly monitoring/debugging/logging/analyzing your applications.
The Monitoring Tool is available as a bookmark at various component levels, such as application, pods, and container.

## Use case

To monitor/debug an application using a specific Monitoring Tool (such as Grafana, Kibana, etc.), you may need to navigate to the tool's page, then to the respective app/resource page.

External links take you directly to the tool's page, which includes the context of the application, environment, pod, and container.

## Prerequisites

Before you begin, configure an application in the Devtron dashboard.

- Super admin access*
- Monitoring tool URL

<sup>*</sup>External links can only be added/managed by a super admin, but other users can [access the configured Monitoring tools](././../creating-application/app-details.md) on their app's page.

## Add an external link

1. On the Devtron dashboard, select `Global Configurations` from the left navigation pane.
2. Select `External links`.
   
![External links welcome page](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/external-links-welcome.png)

3. Select **Add link**.
4. On the `Add link` page, enter the following fields:

![Create an external link](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/add-external-link.png)

    <table>
    <row>
        <th>Field name</th>
        <th>Description</th>
    </row>
    <tr>
        <td><b>Monitoring Tool</b></td>
        <td>Select a Monitoring Tool from the drop-down list. To add a different tool, select 'Other'.</td>
    </tr>
    <tr>
        <td><b>Name</b></td>
        <td>Enter a user-defined name for the Monitoring Tool</td>
    </tr>
    <tr>
        <td><b>Clusters</b></td>
        <td>
            Choose the clusters for which you want to configure the selected tool.
            <ul>
                <li>Select more than one cluster name, to enable the link on multiple clusters</li>
                <li>Select 'Cluster: All', to enable the link on the existing clusters and future clusters</li>                
            </ul>
        </td>
    </tr>
    <tr>
        <td><b>URL Template</b></td>
        <td>
            The configured URL Template is used by apps deployed on the selected clusters.            
            By combining one or more of the env variables, a URL with the structure shown below can be created:<br></br>
            <i>http://www.domain.com/{namespace}/{appName}/details/{appId}/env/{envId}/details/{podName}</i>
            <br></br>
            The env variables:
            <ul>
                <li>{appName}</li>
                <li>{appId}</li>
                <li>{envId}</li>
                <li>{namespace}</li>
                <li>{podName}</li>
                <li>{containerName}</li>
            </ul>
            <b>Note: URL template is dynamically generated from the env variables provided at the time of adding the link.</b><br></br>
            For example:
            <code>https://www.grafana.com/grafana/devtroncd/demo-app/details/24/prod/191/details/my-pod-name</code>
        </td>
    </tr>
    </table>

    > Tip: To add multiple links, select **+ Add another** at the top-left corner.

5. Select **Save**.

## Access an external link

The users (admin and others) can [access the configured external link](././../creating-application/app-details.md) from the **App details** page.

## Manage external links

On this page, the configured external links can be filtered/searched, as well as edited/deleted.

1. Select `Global Configurations > External links`.

![Manage external links](https://devtron-public-asset.s3.us-east-2.amazonaws.com/external-tools/manage-external-links.png)

* Filter and search the links based on the tool's name or a user-defined name.
* Edit a link by selecting the edit icon next to an external link.
* Delete an external link by selecting the delete icon next to a link. The bookmarked link will be removed in the clusters for which it was configured.
