# Comptes Git
Les comptes Git vous permettent de connecter votre source de code à Devtron. Vous pourrez utiliser ces comptes git pour générer le code à l'aide du CI pipeline.
## Ajouter un compte Git
Pour ajouter un compte git, allez dans la section `Comptes Git` des `Configurations générales`. Cliquez sur **Ajouter un compte git**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/git-accounts.jpg)

Renseignez les informations dans les champs suivants pour ajouter votre compte git :

|Champ|Description|
| :- | :- |
|`Nom`|Indiquez un nom à votre fournisseur Git.<br>Note : Ce nom sera disponible dans la Configuration de l'app > Liste déroulante de [Git repository](../creating-application/git-material.md).</br>|
|`Hôte Git`|Il s'agit du fournisseur git sur lequel le repository git de l'application correspondante est hébergé.<br>Note : Par défaut, `Bitbucket` et `GitHub` sont disponibles dans la liste déroulante. Vous pouvez en ajouter autant que vous voulez en cliquant sur `[+ Add Git Host]`.</br>|
|`URL`|Fournissez l'`URL `de l'hôte Git.<br>Par exemple : <https://github.com> pour Github, <https://gitlab.com> pour GitLab etc.|
|`Type d'authentification`|Devtron prend en charge trois types d'authentifications :<ul><li>**User auth** : Si vous sélectionnez `User auth` comme type d'authentification, alors vous devez fournir le `Nom d'utilisateur` et le `Mot de passe` ou `Auth token` pour l'authentification de votre compte de contrôle de version.</li></ul> <ul><li>**Anonymous** : Si vous sélectionnez `Anonymous` comme type d'authentification, alors vous n'avez pas besoin de fournir le `Nom d'utilisateur` et le `Mot de passe`.<br>Note : Si le type d'authentification est défini comme `Anonymous`, seul le repository git public sera accessible.</li></ul><ul><li>**SSH Key** : Si vous choisissez `SSH Key` comme type d'authentification, alors vous devez fournir la `Private SSH Key` correspondant à la clé publique ajoutée dans votre compte de contrôle de version.</li></ul>|

## Mise à jour du compte Git
Pour mettre à jour le compte git :

1. Cliquez sur le compte git que vous voulez mettre à jour.
1. Mettez à jour les modifications requises.
1. Cliquez sur `Mettre à jour` pour sauvegarder les modifications.

Les mises à jour ne peuvent être effectuées qu'au sein d'un seul type d'authentification ou d'un seul type de protocole, c'est-à-dire HTTPS (Anonymous ou User Auth) & SSH. Vous pouvez effectuer une mise à jour d'`Anonymous `vers `User Auth` et vice versa, mais pas d'`Anonymous` ou `User Auth` vers `SSH` et vice versa.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/update-git-accounts.jpg)

Note :

* Vous pouvez activer ou désactiver un compte git. Les comptes git activés seront disponibles dans la Configuration de l'app> [Git repository](../creating-application/git-material.md).

![](../../user-guide/global-configurations/images/git-account-enable-disable.jpg)
