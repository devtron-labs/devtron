# User Access

{% embed url="https://www.youtube.com/watch?v=VTqBRIFbuSU" caption="" %}

Like any enterprise product, Devtron supports fine grained access control to the resources based on

1. Type of action allowed on the Devtron resources (Create Vs View)
2. Sensitivity of the data (Editing image Vs Editing memory)

Access can be added to the User either directly or via Groups.

## Role-based Access Levels

Devtron supports 5 levels of access:

1. **View**: Users with `view` access have the least privileges. Such users can only view combination of environments, applications and helm charts on which access has been granted to the user. They cannot view sensitive data like secrets used in applications or charts.
2. **Build and Deploy**: In addition to `view` privilege mentioned above, users with `build and deploy` permission can build and deploy the image of permitted applications and helm charts to permitted environments.
3. **Admin**: Users with `admin` privileges can create, edit, delete, and view permitted applications in permitted projects.
4. **Manager**: Users with `manager` privileges can do everything that an `admin` user can do. Additionally, they can also give and revoke access of users for the applications and environments of which they are the manager.
5. **Super Admin**: Users with `super admin` privileges have unrestricted access to all the Devtron resources. Super Admins can create, modify, delete and view any Devtron resource without any restriction; it's like Superman without the weakness of Kryptonite. Moreover, they can add and delete user access across any Devtron resource, add delete git repository credentials, container registry credentials, cluster, and environment.
6. **Image approver**: Users with `Image approver` privileges have the authority to approve requests for image deployment.
7. **Configuration approver**: Users with `Configuration approver` privileges have the authority to approve changes to protected configurations of Deployment Template, ConfigMaps, and Secrets. However, the user who proposed the changes cannot self-approve, even if they have configuration approver or super-admin privileges.
8. **Artifact promoter**: Users with this privilege have the authority to approve promotion of [artifacts](../../reference/glossary.md#artifacts) directly to the target CD pipeline.


## User Roles And Permissions

### 1. Custom Applications

| User Roles     | View       | Create     | Edit     | Delete     | Build & Deploy|
| :---           | :---:      | :---:      | :---:    | :---:      | :---:         |
| View             | Yes      | No         | No       | No         | No            |
| Build and Deploy | Yes      | No         | No       | No         | Yes           |
| Admin            | Yes      | Yes        | Yes      | Yes        | Yes           |
| Manager          | Yes      | Yes        | Yes      | Yes        | Yes           |
| Super Admin      | Yes      | Yes        | Yes      | Yes        | Yes           |

### 2. Helm Charts

| User Roles       | View        | Deploy         | Edit        | Delete        |
| :---:            |  :---:      |    :---:       |   :---:     |   :---:       |
| View Only        | Yes         | No             | No          | No            |
| Build and Deploy | Yes         | No             | No          | No            |
| Admin            | Yes         | Yes            | Yes         | Yes           |
| Manager          | Yes         | Yes            | Yes         | Yes           |
| Super Admin      | Yes         | Yes            | Yes         | Yes           |

#### 3. User Access

| User Roles   | Add User Access | Edit User Access | Delete User Access |
| :---         | :---:           | :---:            | :---:              |
| Manager      | Yes             | Yes              | Yes                |
| Super Admin  | Yes             | Yes              | Yes                |

#### 4. Global Configurations

| User Role    | Add Global Config | Edit Global Config | Delete Global Config |
| :---         | :---:             | :---:              | :---:                |
| Super Admin  | Yes               | Yes                | Yes                  |

To control the access of User and Group-

Go to the left main panel → Global Configurations` → `User Access`

## Users

### 1. Add new user

Click **Add User**, to add one or multiple users.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/gc-user-access-add-user.jpg)

### 2. Create User Permissions

When you click `Add User`, you will see 6 options to set permission for users which are as follow:

- Email addresses
- Assign super admin permissions
- Group Permissions
- Devtron Apps Permissions
  - Project
  - Environment
  - Applications
  - Roles
- Helm Apps Permissions
  - Project
  - Environment or cluster/namespace
  - Applications
  - Permission
- Chart group permissions

## Email addresses:

In the `Email address` box, you have to provide the mail ID of the user to whom you want to give access to your applications.

**`IMP`** Please note that Email address should be same as that in the `email` field in the JWT token returned by OIDC provider.

### Assign super admin permissions

If you check the option `Assign super admin permissions`, the user will get full access to your system and the rest of the options will disappear. Please check [above](#role-based-access-levels) to see permission levels. Only users with super admin permissions can assign super admin permissions to a user.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/add-user-2.jpg)

Click **Save** and your user will be saved with super admin permissions.

We suggest that super admin privileges should be given to only select few.

If you don’t want to assign super admin permissions then you have to provide the rest of the information.

### Group Permissions

This is used to assign user to a particular group and user inherits all the permissions granted to this group. The Group permissions section contains a drop-down of all existing groups on which you have access. This is optional field and more than one groups can be selected for a user.

We will discuss how to create groups in the later section.

### Devtron Apps permissions

Access to devtron applications can be given to user by attaching permission directly to his/her email id through the `Devtron Apps` section. This section has 4 options to manage the permissions of your users.

- **Project**

Select a project from the drop-down to which you want to give permission to the users. You can select only one project at a time if you want to select more than one project then click **Add row**.

- **Environment**

In the `Environment` section, you can select one or more than one or all environments at a time. Click on the environment section, you will see a drop-down of your environments and select any environment on which you want to give permission to the user.

**`IMP`** If `all environments` option is selected then user gets access to all current environments and any new environment which gets associated with this application later.

- **Applications**

Similarly, you can select `Applications` from the drop-down corresponding to your selected Environments. In this section, you can also give permissions to one or more than one or to all applications at a time.

**`IMP`** If `all applications` option is selected then user gets access to all current applications and any new application which gets associated with this project later.

- **Roles**

  Inside the `Role`, you actually choose which type of permissions you want to give to the users.

There are four different view access levels/Role available for both User and Group as described [above](#role-based-access-levels):

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/user-access-permission_1.JPG)

You can add multiple rows, for Devtron app permission.

Once you have finished assigning the appropriate permissions for the listed users, click **Save**.

### Helm Apps Permissions

Access to devtron applications can be given to user by attaching permission directly to his/her email id through the `Devtron Apps` section. This section has 4 options to manage the permissions of your users.

- **Project**

Select a project from the drop-down to which you want to give permission to the users. You can select only one project at a time if you want to select more than one project then click **Add row**.

- **Environment or cluster/namespace**

In the `Environment` section, you can select one or more than one or all environments at a time. Click on the environment section, you will see a drop-down of your environments and select any environment on which you want to give permission to the user.

**`IMP`** If `all environments` option is selected then user gets access to all current environments and any new environment which gets associated with this application later.

- **Applications**

Similarly, you can select `Applications` from the drop-down corresponding to your selected Environments. In this section, you can also give permissions to one or more than one or to all applications at a time.

**`IMP`** If `all applications` option is selected then user gets access to all current applications and any new application which gets associated with this project later.

- **Permission**

  Inside the `Role`, you actually choose which type of permissions you want to give to the users.

There are four different view access levels/Role available for both User and Group as described [above](#role-based-access-levels):

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/user-access-permission_helm+app_2.JPG)

### Chart Group Permissions

You can also manage the access of users to Chart Groups in your project.

**NOTE:** You can only give users the ability to `create` or `edit`, not both.

Click on the checkbox of `Create`, if you want the users to create, view, edit, or delete the chart groups.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/chart+group+permission_create_3.JPG)

To permit a user to only `edit` the chart groups, check `Specific chart group` from `Edit` drop-down. In the following field, select the chart group for which you want to grant the user edit permission.

Go to `Edit` drop-down, if you want to `allow` or `deny` users to edit the chart groups.

Select on `Deny` option from the drop-down, if you want to restrict the users to edit the chart groups.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/chart+group+permission_edit_4.JPG)

Select the `Specific Charts` option from the drop-down and then select the chart groups for which you want to allow users to edit, from the other drop-down menu.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/chart+group+permission_edit+specific+chart_5.JPG)

Click on `Save`, once you have configured all the required permissions for the users.

| Action | Permissions |
| :---   | :---         |
| View  | Only can view chart groups |
| Create | Can create, view, edit or delete |
| Edit | <ul><br/><li>**Deny:** Can't edit chart groups</li><li>**Specific chart groups:** can edit specific chart group</li></ul> |

### 3. Edit User Permissions

You can edit the user permissions, by clicking the `downward arrow`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/edit+user+permission_dropdown_1.JPG)

Then you can edit the user permissions here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/edit+user+permission_devtron+app_2.JPG)

After you have done editing the user permissions, click **Save**.

If you want to delete the user/users with particular permissions, click **Delete**.

## Groups

The advantage of the groups is to define a set of privileges like create, edit, or delete for the given set of resources that can be shared among the users within the group. Users can be added to an existing group to utilize the privileges that it grants. Any access change to group is reflected immediately in user access.

You can select the group which you are creating in the `Group permissions` section inside `Add users`.

### 1. Add new Group

Go to `Groups` and click **Add Group**, to create a new group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/Add+group+permission_dashboard_1.JPG)

Enter the `Group Name` and `Description`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/group+permission_description_2.JPG)

### 2. Create Group Permissions

Once you have given the group name and group description.

Then, control the access permissions of groups in the Devtron Apps, Helm Apps or Group Chart Permissions section. Manage the Project, Environment, Application, and Role access the same as we discuss in the above users section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/group+permission_devtron+app_3.JPG)

You can add multiple rows, for the Devtron Apps and Helm Apps Permissions section.

Once you have finished assigning the appropriate permissions for the listed users, Click **Save**.

### 3. Edit Group Permissions

You can edit the group permissions, by clicking the `downward arrow.`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/group+permission_dropdown_4.JPG)

Then you can edit the user permissions here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/user-access/edit+group+permission_devtron+app_5.JPG)

Once you are done editing the group permissions. Click **Save**.

If you want to delete the groups with particular permissions. Click **Delete**.

### 4. Manage Chart Group Permissions

The chart group permissions for the group will be managed in the same way as for the users. For reference, check Manage chart group permissions for users.
