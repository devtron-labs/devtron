# App Labels Tagging

Overview
============

This feature helps you to tag labels on applications, this will help you to filter out and categories your applications.

labels are combinations of key value pairs, each application may have multiple labels.

every app must have unique key.

### 1. Add labels on creation of application

Login with valid credentials and go to `Applications` and click on right top button `Add New App` and fill the required
values.

Labels are optionals and can be entered `key:values` format. multiple labels can be added without repeating `key` name.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/app-labels/app-labels-1.jpg)

### 2. Application meta info

Login with valid credentials and go to `Applications` and click on any application and go to detail page, click on top
left button next to application name `?`.

This will open show you the applications meta details like project, labels associated with it.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/app-labels/app-labels-2.jpg)

### 3. Edit Labels to existing apps

We also add or remove labels for app from here.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/app-labels/app-labels-2.jpg)

### 3. Label Payload

* `id` : integer unique label id
* `appId` : integer application id
* `key` : string key is the part of label stored individually in db.
* `value` : string value is the part of label stored individually in db.
