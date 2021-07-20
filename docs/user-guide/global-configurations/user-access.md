
# User Access

{% embed url="https://www.youtube.com/watch?v=VTqBRIFbuSU" caption="" %}

Like any entreprise product, Devtron supports fine grained access control to the resources based on
1. Type of action allowed on the Devtron resources (Create Vs View)
2. Sensitivity of the data (Editing image Vs Editing memory)

Access can be added to the User either directly or via Groups.

## Access Levels
Devtron supports 5 levels of access
1. **View**: User with `view` only access has the least privilege. This user can only view combination of environments, applications and helm charts on which access has been granted to the user. This user cannot view sensitive data like secrets used in applications or charts.
2. **Build and Deploy**: In addition to `view` privilege mentioned in above, user with `build and deploy` permission can build and deploy the image of permitted applications and helm charts to permitted environments.
3. **Admin**: User with `admin` access can create, edit, delete and view permitted applications in permitted projects.
4. **Manager**: User with `manager` access can do everything that an `admin` type user can do, in addition they can also give and revoke access of users for the applications and environments of which they are `manager`.
5. **Super Admin**: User with `super admin` privilege has unrestricted access to all Devtron resources. Super admin can create, modify, delete and view any Devtron resource without any restriction; its like Superman without the weakness of Kryptonite. Super Admin can also add and delete user access across any Devtron resource, add delete git repository credentials, docker registry credentials, cluster and environment.

| Access Level | View App | Create App | Edit App | Delete App | Trigger App |
|--|--|--|--|--|--|
| View | Yes  | No | No | No| No |
| Build and Deploy | Yes  | No | No | No| Yes |
|Admin| Yes | Yes | Yes | Yes | Yes |
|Manager| Yes | Yes | Yes | Yes | Yes |
|Super Admin| Yes | Yes | Yes | Yes | Yes |
<br/>

| Access Level | View Charts | Install Charts | Edit Charts | Delete Charts |
|--|--|--|--|--|--|
| View | Yes  | No | No | No|
| Build and Deploy | Yes  | No | No | No|
|Admin| Yes | Yes | Yes | Yes |
|Manager| Yes | Yes | Yes | Yes |
|Super Admin| Yes | Yes | Yes | Yes |
<br/>

| Access Level | Add User Access  | Edit User Access | Delete User Access |
|--|--|--|--|
|Manager| Yes | Yes | Yes |
|Super Admin| Yes | Yes | Yes |
<br/>

| Access Level | Add Global Config  | Edit Global Config | Delete Global Config |
|--|--|--|--|
|Super Admin| Yes | Yes |

To control the access of User and Group-

Go to the left main panel -&gt; `Select Global Configurations` -&gt; Select `User Access`

## Users

### 1. Add new user

Click on `Add User`, to add one or multiple users.

![](../../.gitbook/assets/user1.jpg)

### 2. Create User Permissions

When you click on Add User, you will see 5 options to set permission for users which are as follow:

* Email addresses
* Assign super admin permissions
* Group permissions
* Direct permissions
  * Project
  * Environment
  * Applications
  * Roles
* Chart group permissions

## Email addresses:

In the `Email address` box, you have to provide the mail ID of the user to whom you want to give access to your applications.

**`IMP`** Please note that Email address should be same as that in the `email` field in the JWT token returned by OIDC provider.

### Assign super admin permissions

If you check the option `Assign super admin permissions`, the user will get full access to your system and the rest of the options will disappear.  Please check [above](#access-levels) to see permission levels.

![](../../.gitbook/assets/user_gc1%20(3)%20(4).png)

Click on `Save` and your user will be saved with super admin permissions.

We suggest that super admin privileges should be given to only select few.

If you don’t want to assign super admin permissions then you have to provide the rest of the information.

### Group permissions

This is used to assign user to a particular group and user inherits all the permissions granted to this group. The Group permissions section contains a drop-down of all existing groups on which you have access. This is optional field and more than one groups can be selected for a user.

We will discuss how to create groups in the later section.

### Direct permissions

Access can be given to user by attaching permission directly to his/her email id through the `Direct Permissions` section. This section has 4 options to manage the permissions of your users.

* **Project**

Select a project from the drop-down to which you want to give permission to the users. You can select only one project at a time if you want to select more than one project then click `Add row`.

* **Environment**

In the `Environment` section, you can select one or more than one or all environments at a time. Click on the environment section, you will see a drop-down of your environments and select any environment on which you want to give permission to the user.

**`IMP`** If `all environments` option is selected then user gets access to all current environments and any new environment which gets associated with this application later.

* **Applications**

Similarly, you can select `Applications` from the drop-down corresponding to your selected Environments. In this section, you can also give permissions to one or more than one or to all applications at a time.

**`IMP`** If `all applications` option is selected then user gets access to all current applications and any new application which gets associated with this project later.

* **Roles**

  Inside the `Role`, you actually choose which type of permissions you want to give to the users.

There are four different view access levels/Role available for both User and Group as described [above](#access-levels):

![](../../.gitbook/assets/gc-user-access-configure-direct-permission%20(3)%20(6).jpg)

You can add multiple rows, for Direct Permissions.

Once you have finished assigning the appropriate permissions for the listed users, Click on `Save`.

### Chart group permissions

You can also manage the access of users to Chart Groups in your project. By default, user has the `View` permission for the charts deployed in the environments on which he/she has at least `view` access to any application.

You can either give the users permission to `Create` or `Edit`.

Click on the checkbox of `Create`, if you want the users to create, view, edit, or delete all the chart groups.

![](../../.gitbook/assets/gc-user-access-configure-chart-permission%20(1)%20(3)%20(4).jpg)

Click on the checkbox of `Edit`, if you want to `allow` or `deny` users to edit the chart groups.

Select on `Deny` option from the drop-down menu, if you want to restrict the users to edit the chart groups.

![](../../.gitbook/assets/user5%20(2).jpg)

Select the `Specific Charts` option from the drop-down menu, and then select the chart groups for which you want to allow users to edit, from the other drop-down menu.

![](../../.gitbook/assets/user6%20(1)%20(6).jpg)

Click on `Save`, once you have configured all the required permissions for the users.

### 3. Edit User Permissions

You can edit the user permissions, by clicking on the `downward arrow`.

![](../../.gitbook/assets/gc-user-access-edit-arrow%20%282%29.jpg)

Then you can edit the user permissions here.

![](../../.gitbook/assets/gc-user-access-edit-user-permission%20(6).jpg)

After you have done editing the user permissions. Click on `Save`.

If you want to delete the user/users with particular permissions. Click on `Delete`.

## Groups

The advantage of the groups is to define a set of privileges like create, edit, or delete for the given set of resources that can be shared among the users within the group. Users can be added to an existing group to utilize the privileges that it grants. Any access change to group is reflected immediately in user access.

You can select the group which you are creating in the `Group permissions` section inside `Add users`.

### 1. Add new Group

Click on `Add Group`, to create a new group.

![](../../.gitbook/assets/user7%20(2)%20(5).jpg)

Enter the `Group Name` and `Description`.

![](../../.gitbook/assets/gc-user-access-add-group-name%20(2).jpg)

### 2. Create Group Permissions

Once you have given the group name and group description.

Then, control the access permissions of groups in the Direct Permissions section. Manage the Project, Environment, Application, and Role access the same as we discuss in the above users section.

![](../../.gitbook/assets/gc-user-access-add-group-configure%20(3).jpg)

You can add multiple rows, for the Direct Permissions section.

Once you have finished assigning the appropriate permissions for the listed users, Click on `Save`.

### 3. Edit Group Permissions

You can edit the group permissions, by clicking on the `downward arrow.`

![](../../.gitbook/assets/gc-user-access-add-group-edit-arrow.jpg)

Then you can edit the user permissions here.

![](../../.gitbook/assets/gc-user-access-add-group-edit-configure%20(3).jpg)

Once you are done editing the group permissions. Click on `Save`.

If you want to delete the groups with particular permissions. Click on `Delete`.

### 4. Manage Chart Group Permissions

The chart group permissions for the group will be managed in the same way as for the users. For reference, check Manage chart group permissions for users.