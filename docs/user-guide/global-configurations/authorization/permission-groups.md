# Permission Groups

This is used to assign user to a particular group and user inherits all the permissions granted to this group. The `Permission groups` section contains a drop-down of all existing groups on which you have access. This is optional field and more than one groups can be selected for a user.


The advantage of the groups is to define a set of privileges like create, edit, or delete for the given set of resources that can be shared among the users within the group. Users can be added to an existing group to utilize the privileges that it grants. Any access change to group is reflected immediately in user access.

You can select the group which you are creating in the `Group permissions` section inside `Add users`.

## 1. Add new Group

Go to `Global configurations` -> `Authorization` -> `Permission group` and click on `Add Group` to create a new group.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-1.png)

Enter the `Group Name` and `Description`.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-2.png)

## 2. Create permission Group

Once you have given the group name and group description.

Assign the permissions of groups in the `Devtron Apps`, `Helm Apps` or `Group Chart` permissions section. Manage the project, environment, application and role access the same as we discuss in the user permissions section.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-3.png)

You can add multiple rows for the `Devtron Apps` and `Helm Apps` Permissions section.

Once you have finished assigning the appropriate permissions for the permission group, click on `Save`.

## 3. Edit Group Permissions

You can edit the permission groups by clicking on the `downward arrow.`

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-4.png)

Then you can edit the permission group here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/permission-group/permission-group-5.png)

Once you are done editing the permission group, click on `Save`.

If you want to delete the groups with particular permission group, click on `Delete`.

## 4. Manage Chart Group Permissions

The chart group permissions for the permission groups will be managed in the same way as for the users. For reference, check Manage chart group permissions for users.
