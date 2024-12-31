# Git Repository

## Introduction

During the [CI process](../deploying-application/triggering-ci.md), the application source code is pulled from your [git repository](../../reference/glossary.md#repo). 

Devtron also supports multiple Git repositories (be it from one Git account or multiple Git accounts) in a single deployment.

![Figure 1: Adding Git Repository](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/add-git-repo.jpg)

Therefore, this doc is divided into 2 sections, read the one that caters to your application:
* [Single Repo Application](#single-repo-application)
* [Multi Repo Application](#multi-repo-application)

---

## Single Repo Application

Follow the below steps if the source code of your application is hosted on a single Git repository.

In your application, go to **App Configuration** → **Git Repository**. You will get the following fields and options:

1. [Git Account](#git-account)
2. [Git Repo URL](#git-repo-url)
3. (Checkboxes)
    * [Exclude specific file/folder in this repo](#exclude-specific-filefolder-in-this-repo)
    * [Set clone directory](#set-clone-directory)
    * [Pull submodules recursively](#pull-submodules-recursively)

### Git Account

This is a dropdown that shows the list of Git accounts added to your organization on Devtron. If you haven't done already, we recommend you to first [add your Git account](../global-configurations/git-accounts.md) (especially when the repository is private).

![Figure 2: Selecting Git Account](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/select-git-account.jpg)

{% hint style="info" %}
If the authentication type of your Git account is anonymous, only public Git repositories in that account will be accessible. Whereas, adding a user auth or SSH key will make both public and private repositories accessible.
{% endhint %}


### Git Repo URL

In this field, you have to provide your code repository’s URL, for e.g., `https://github.com/devtron-labs/django-repo`.

You can find this URL by clicking on the **Code** button available on your repository page as shown below:

![Figure 3: Getting Repo URL](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/repo-url.jpg)

{% hint style="info" %}
* Copy the HTTPS/SSH portion of the URL too
* Make sure you've added your [Dockerfile](https://docs.docker.com/engine/reference/builder/) in the repo
{% endhint %}


### Exclude specific file/folder in this repo

Not all repository changes are worth triggering a new [CI build](../deploying-application/triggering-ci.md). If you enable this checkbox, you can define the file(s) or folder(s) whose commits you wish to use in the CI build.

![Figure 4: Sample Exclusion Rule](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/sample1.jpg)

In other words, if a given commit contains changes only in file(s) present in your exclusion rule, the commit won't show up while selecting the [Git material](../../reference/glossary.md#material), which means it will not be eligible for build. However, if a given commit contains changes in other files too (along with the excluded file), the commit won't be excluded and it will definitely show up in the list of commits.

![Figure 5: Excludes commits made to README.md](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/excluded-commit.jpg)

Devtron allows you to create either an exclusion rule, an inclusion rule, or a combination of both. In case of multiple files or folders, you can list them in new lines. 

To exclude a path, use **!** as the prefix, e.g. `!path/to/file` <br />
To include a path, don't use any prefix, e.g. `path/to/file`


#### Examples


| Sample Values | Description |
|---|---|
| `!README.md` | **Exclusion of a single file in root folder:** <br/> Commits containing changes made only in README.md file will not be shown |
| `!README.md` <br /> `!index.js` | **Exclusion of multiple files in root folder:** <br/> Commits containing changes made only in README.md or/and index.js files will not be shown |
|  `README.md` | **Inclusion of a single file in root folder:** <br/> Commits containing changes made only in README.md file will be shown. Rest all will be excluded. |
|  `!src/extensions/printer/code2.py` | **Exclusion of a single file in a folder tree:** <br/> Commits containing changes made specifically to code2.py file will not be shown |
|  `!src/*` | **Exclusion of a single folder and all its files:** <br/> Commits containing changes made specifically to files within src folder will not be shown |
|  `!README.md` <br/> `index.js` | **Exclusion and inclusion of files:** <br/> Commits containing changes made only in README.md will not be shown, but commits made in index.js file will be shown. All other commits apart from the aforementioned files will be excluded. |
|  `!README.md` <br/> `README.md` | **Exclusion and inclusion of conflicting files:** <br/> If conflicting paths are defined in the rule, the one defined later will be considered. In this case, commits containing changes made only in README.md will be shown. |


You may use the **Learn how** link (as shown below) to understand the syntax of defining an exclusion or inclusion rule.

![Figure 6: 'Learn how' Button](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/rules.jpg)

Since file paths can be long, Devtron supports regex too for writing the paths. To understand it better, you may click the **How to use** link as shown below.

![Figure 7: Regex Support](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/regex-help.jpg)

#### How to view excluded commits?

As we saw earlier in fig. 4 and 5, commits containing the changes of only `README.md` file were not displayed, since the file was in the exclusion list. 

However, Devtron gives you the option to view the excluded commits too. There's a döner menu at the top-right (beside the `Search by commit hash` search bar).

![Figure 8a: Döner Menu Icon](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/doner-menu.jpg)

![Figure 8b: Show Excluded Commits](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/show-exclusions.jpg)

![Figure 8c: Commits Unavailable for Build](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/excluded-commits.jpg)

The **EXCLUDED** label (in red) indicates that the commits contain changes made only to the excluded file, and hence they are unavailable for build.


### Set clone directory

After clicking the checkbox, a field titled `clone directory path` appears. It is the directory where your code will be cloned for the repository you specified in the previous step.

This field is optional for a single Git repository application and you can leave the path as default. Devtron assigns a directory by itself when the field is left blank. The default value of this field is `./`

![Figure 8: Clone Directory Option](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/git-material/clone-directory.jpg)


### Pull submodules recursively

This checkbox is optional and is used for pulling [git submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules) present in a repo. The submodules will be pulled recursively, and the auth method used for the parent repo will be used for submodules too.

---

## Multi Repo Application

As discussed earlier, Devtron also supports multiple git repositories in a single application. To add multiple repositories, click **Add Git Repository** and repeat all the steps as mentioned in [Single Repo Application](#single-repo-application). However, ensure that the clone directory paths are unique for each repo. 

Repeat the process for every new git repository you add. The clone directory path is used by Devtron to assign a directory to each of your Git repositories. Devtron will clone your code at those locations and those paths can be referenced in the Docker file to create a Docker image of the application.

Whenever a change is pushed to any of the configured repositories, CI will be triggered and a new Docker image file will be built (based on the latest commits of the configured repositories). Next, the image will be pushed to the container registry you configured in Devtron.

{% hint style="info" %}
Even if you add multiple repositories, only one image will be created based on the Dockerfile as shown in the [docker build config](docker-build-configuration.md)
{% endhint %}

### Why do you need Multi-Git support?

Let’s look at this with an example:

Due to security reasons, you want to keep sensitive configurations like third-party API keys in separate access-restricted git repositories, and the source code in a Git repository that every developer has access to. To deploy this application, code from both the repositories are required. A Multi-Git support helps you achieve it.

Other examples where you might need Multi-Git support:

* To make code modularized, where front-end and back-end code are in different repos
* Common library extracted out in a different repo so that other projects can use it