# User Access for Hyperion Mode

Like any enterprise product, Devtron supports fine grained access control to the resources

Access can be added to the User either directly or via Groups.

## Access Levels
Devtron supports 4 levels of access
1. **View only**: User with `view` only access has the least privilege. This user can only view combination of environments, applications and helm charts on which access has been granted to the user. This user cannot view sensitive data like secrets used in applications or charts.
2. **View and Edit**: In addition to `view` privilege mentioned in above, user with `View and Edit` permission can edit the resource manifests of permitted applications and helm charts to permitted environments.
3. **Admin**: User with `admin` access can create, edit, delete and view permitted applications in permitted projects.
4. **Super Admin**: User with `super admin` privilege has unrestricted access to all Devtron resources. Super admin can create, modify, delete and view any Devtron resource without any restriction; its like Superman without the weakness of Kryptonite. Super Admin can also add and delete user access across any Devtron resource, add delete git repository credentials, docker registry credentials, cluster and environment.

## Visualize using access table (Apps)

| Access Level | View App | Create App | Edit App | Delete App | Trigger App |
|--|--|--|--|--|--|
| View | Yes  | No | No | No| No |
| View and Edit | Yes  | No | Yes | Yes| Yes |
|Admin| Yes | Yes | Yes | Yes | Yes |
|Super Admin| Yes | Yes | Yes | Yes | Yes |
<br/>

## Visualize using access table (Charts)
| Access Level | View Charts | Install Charts | Edit Charts | Delete Charts |
|--|--|--|--|--|
| View | Yes  | No | No | No|
| View and Edit | Yes  | No | No | No|
|Admin| Yes | Yes | Yes | Yes |
|Super Admin| Yes | Yes | Yes | Yes |
<br/>

## Visualize using access table (User Management)
| Access Level | Add User Access  | Edit User Access | Delete User Access |
|--|--|--|--|
|Super Admin| Yes | Yes | Yes |
<br/>

## Visualize using access table (Config Management)
| Access Level | Add Global Config  | Edit Global Config | Delete Global Config |
|--|--|--|--|
|Super Admin| Yes | Yes |


To control the access of User and Group

Go to the left main panel -&gt; `Select Global Configurations` -&gt; Select `User Access`

## Users
### 1. Add new user

Click on `Add User`, to add one or multiple users.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-add-user.jpg)

### 2. Search the existing User
Click on `Search Box`, and type your user's email

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-search-user.jpg)

### 3. Create User Permissions

When you click on Add User, you will see 4 options to set permission for users which are as follow:

* Email addresses
* Assign super admin permissions
* Group permissions
* Helm Apps
  * Project
  * Environment
  * Application
  * Permission
  
## Email addresses:

In the `Email address` box, you have to provide the mail ID of the user to whom you want to give access to your applications.

**`IMP`** Please note that Email address should be same as that in the `email` field in the JWT token returned by OIDC provider.


### Assign super admin permissions

If you check the option `Assign super admin permissions`, the user will get full access to your system and the rest of the options will disappear.  Please check [above](#access-levels) to see permission levels.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-assign-superuser.jpg)

Click on `Save` and your user will be saved with super admin permissions.

We suggest that super admin privileges should be given to only select few.

If you don’t want to assign super admin permissions then you have to provide the rest of the information.


### Group permissions

This is used to assign user to a particular group and user inherits all the permissions granted to this group. The Group permissions section contains a drop-down of all existing groups on which you have access. This is optional field and more than one groups can be selected for a user.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-assign-groups-user.jpg)

We will discuss how to create groups in the later section.

### Helm Apps

Access can be given to user by attaching permission directly to his/her email id through the `Helm Apps` section. This section has 4 options to manage the permissions of your users.

* **Project**

Select a project from the drop-down to which you want to give permission to the users. You can select only one project at a time if you want to select more than one project then click `Add row`.

* **Environment or cluster/namespace**

In the `Environment or cluster/namespace` section, you can select one or more than one or all environments at a time. Click on the environment section, you will see a drop-down of your environments and select any environment on which you want to give permission to the user.

**`IMP`** If `all environments` option is selected then user gets access to all current environments and any new environment which gets associated with this application later.

* **Application**

Similarly, you can select `Applications` from the drop-down corresponding to your selected Environments. In this section, you can also give permissions to one or more than one or to all applications at a time.

**`IMP`** If `all applications` option is selected then user gets access to all current applications and any new application which gets associated with this project later.

* **Permission**

  Inside the `Permission`, you actually choose which type of permissions you want to give to the users.

There are there different view access levels available for both User and Group as described [above](#access-levels):

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-user-permission.jpg)

You can add multiple rows, for Helm Apps.

Once you have finished assigning the appropriate permissions for the listed user, Click on `Save`.

### 4. Edit User Permissions

You can edit the user permissions, by clicking on the `downward arrow`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-user-edit-arrow.jpg)

Then you can edit the user permissions here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-user-edit-permission.jpg)

After you have done editing the user permissions. Click on `Save`.

If you want to delete the user/users with particular permissions. Click on `Delete`.

## Groups

The advantage of the groups is to define a set of privileges like create, edit, or delete for the given set of resources that can be shared among the users within the group. Users can be added to an existing group to utilize the privileges that it grants. Any access change to group is reflected immediately in user access.

You can select the group which you are creating in the `Group permissions` section inside `Add users`.

### 1. Add new Group

Click on `Add Group`, to create a new group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-group.jpg)

Enter the `Group Name` and `Description`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-group-name-desc.jpg)

### 2. Create Group Permissions

Once you have given the group name and group description.

Then, control the access permissions of groups in the Helm Apps section. Manage the Project, Environment, Application, and Permission access the same as we discuss in the above users section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-group-add-permission.jpg)

You can add multiple rows, for the Helm Apps section.

Once you have finished assigning the appropriate permissions for the listed users, Click on `Save`.

### 3. Edit Group Permissions

You can edit the group permissions, by clicking on the `downward arrow.`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-group-edit-arow.jpg)

Then you can edit the group permissions here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/hyperion/user-guide/global-configurations/user-access/hyperion-gc-group-edit-permission.jpg)

Once you are done editing the group permissions. Click on `Save`.

If you want to delete the groups with particular permissions. Click on `Delete`.


### 4. Manage Chart Group Permissions

The chart group permissions for the group will be managed in the same way as for the users. For reference, check Manage chart group permissions for users.


