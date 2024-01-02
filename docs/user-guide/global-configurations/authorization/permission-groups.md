# Permission Groups

Using the `Permission groups`, you can assign a user to a particular group and a user inherits all the permissions granted to the group. 

The advantage of the `Permission groups` is to define a set of privileges like create, edit, or delete for the given set of resources that can be shared among the users within the group.

**Note**: The [User permissions](https://docs.devtron.ai/global-configurations/authorization/user-access) section for `Specific permissions` contains a drop-down list of all existing groups for which a user has an access. This is an optional field and more than one groups can be selected for a user.

## Add Group

To add a group, go to the `Authorization > Permissions groups` section of `Global Configurations`. Click **Add group**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-1.png)

Enter the `Group Name` and `Description`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-2.png)

You can either grant [super-admin](../../global-configurations/authorization/user-access.md#role-based-access-levels) permission to a user group or specific permissions to manage access for:

   * [Devtron Apps](#devtron-apps-permissions)
   * [Helm Apps](#helm-apps-permissions)
   * [Kubernetes Resources](#kubernetes-resources-permissions)
   * [Chart Groups](#chart-group-permissions)

### Devtron Apps Permissions

In `Devtron Apps` option, you can provide access to a group to manage permission for custom apps created using Devtron.

**Note**: The `Devtron Apps` option will be available only if you install [CI/CD integration](https://docs.devtron.ai/usage/integrations/build-and-deploy-ci-cd).

Provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-devtron-apps.jpg)


| Registry Type | Credentials |
| --- | --- |
| **Project** | Select a project from the drop-down list to which you want to give permission to the group. You can select only one project at a time.<br>Note: If you want to select more than one project, then click `Add row`.</br> |
| **Environment** | Select the specific environment or all environments from the drop-down list.<br>Note: If you select `All environments` option, then a user gets access to all the current environments including any new environment which gets associated with the application later.</br> |
| **Application**  | Select the specific applications or all applications from the drop-down list corresponding to your selected Environments.<br>Note: If you select `All applications` option, then a user gets access to all the current applications including any new application which gets associated with the project later</br>.  |
| **Role**  | Select one of the [roles](#role-based-access-levels) to which you want to give permission to the user:<ul><li>`View only`</li></ul> <ul><li>`Build and Deploy`</li></ul><ul><li>`Admin`</li></ul><ul><li>`Manager`</li></ul>  |

You can add multiple rows for `Devtron Apps` permission.

Once you have finished assigning the appropriate permissions for the groups, Click `Save`.

### Helm Apps Permissions

In `Helm Apps` option, you can provide access to a group to manage permission for Helm apps deployed from Devtron or outside Devtron.

Provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-groups-helm-apps.jpg)

| Registry Type | Credentials |
| --- | --- |
| **Project** | Select a project from the drop-down list to which you want to give permission to the group. You can select only one project at a time.<br>Note: If you want to select more than one project, then click `Add row`.</br> |
| **Environment or cluster/namespace** | Select the specific environment or `all existing environments in default cluster` from the drop-down list.<br>Note: If you select `all existing + future environments in default cluster` option, then a user gets access to all the current environments including any new environment which gets associated with the application later.</br> |
| **Application**  | Select the specific application or all applications from the drop-down list corresponding to your selected Environments.<br>Note: If `All applications` option is selected, then a user gets access to all the current applications including any new application which gets associated with the project later</br>.  |
| **Role**  | Select one of the [roles](#role-based-access-levels) to which you want to give permission to the user:<ul><li>`View only`</li></ul> <ul><li>`View & Edit`</li></ul><ul><li>`Admin`</li></ul>  |

You can add multiple rows for Devtron app permission.

Once you have finished assigning the appropriate permissions for the groups, Click `Save`.


### Kubernetes Resources Permissions

In `Kubernetes Resources` option, you can provide permission to view, inspect, manage, and delete resources in your clusters from [Kubernetes Resource Browser](https://docs.devtron.ai/usage/resource-browser) page in Devtron. You can also create resources from the `Kubernetes Resource Browser` page.

**Note**: Only super admin users will be able to see `Kubernetes Resources` tab and provide permission to other users to access `Resource Browser`.

To provide Kubernetes resource permission, click `Add permission`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/kubernetes-resources-permission-group.jpg)

On the `Kubernetes resource permission`, provide the information in the following fields:


![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/user-permission/kubernetes-resource-permission-page-latest.jpg)

| Registry Type | Credentials |
| --- | --- |
| **Cluster** | Select a cluster from the drop-down list to which you want to give permission to the user. You can select only one cluster at a time.<br>Note: To add another cluster, then click `Add another`.</br> |
| **Namespace** | Select the namespace from the drop-down list. |
| **API Group**  | Select the specific API group or `All API groups` from the drop-down list corresponding to the K8s resource.  |
 **Kind**  | Select the kind or `All kind` from the drop-down list corresponding to the K8s resource.  |
  **Resource name**  | Select the resource name or `All resources` from the drop-down list to which you want to give permission to the user. |
| **Role**  | Select one of the [roles](#role-based-access-levels) to which you want to give permission to the user and click `Done`:<ul><li>`View`</li></ul> <ul><li>`Admin`</li></ul>  |

You can add multiple rows for Kubernetes resource permission.

Once you have finished assigning the appropriate permissions for the groups, Click `Save`.

### Chart Group Permissions

In `Chart group permission` option, you can manage the access of groups for Chart Groups in your project.

**Note**: The `Chart group permission` option will be available only if you install [CI/CD integration](https://docs.devtron.ai/usage/integrations/build-and-deploy-ci-cd).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/kubernetes-resources-permission-group.jpg)

**NOTE:** You can only give users the ability to `create` or `edit`, not both.

| Action | Permissions |
| :---   | :---         |
| View  | Enable `View` to view chart groups only. |
| Create | Enable `Create` if you want the users to create, view, edit or delete the chart groups. |
| Edit | <ul><li>**Deny**: Select `Deny` option from the drop-down list to restrict the users to edit the chart groups.</li><li>**Specific chart groups**: Select the `Specific Charts Groups` option from the drop-down list and then select the chart group for which you want to allow users to edit.</li></ul> |

Click `Save`once you have configured all the required permissions for the groups.


### Edit Permissions Groups

You can edit the permission groups by clicking the `downward arrow.`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-4.png)

Edit the permission group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-5.png)

Once you are done editing the permission group, click `Save`.

If you want to delete the groups with particular permission group, click `Delete`.


