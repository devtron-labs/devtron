# GitOps
Devtron utilise GitOps pour stocker les fichiers de configuration Kubernetes des applications.
Pour stocker les fichiers de configuration et l'état souhaité des applications, les identifiants Git doivent être fournis dans **Configurations générales** > **GitOps** dans le tableau de bord Devtron.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

Vous trouverez ci-dessous les fournisseurs Git qui sont disponibles dans Devtron. Sélectionnez l'un des fournisseurs Git (par exemple, GitHub) pour configurer GitOps :

* [GitHub](#github)
* [GitLab](#gitlab)
* [Azure](#azure)
* [BitBucket Cloud](#bitbucket-cloud)

**Note** : Le fournisseur Git que vous sélectionnez pour configurer GitOps aura un impact sur les sections suivantes :

* Modèle de déploiement, [cliquez ici](https://docs.devtron.ai/user-guide/creating-application/deployment-template) pour en savoir plus.
* Graphiques, [cliquez ici](https://docs.devtron.ai/user-guide/deploy-chart) pour en savoir plus.
## GitHub
Si vous sélectionnez `GitHuB` comme fournisseur git, veuillez renseigner les champs suivants pour configurer GitOps :

|Champs|Description|
| :-: | :-: |
|**Hôte Git**|Ce champ indique l'URL du fournisseur Git sélectionné. <br>A titre d'exemple : https://github.com/ pour GitHub.</br>|
|**Nom de l'organisation GitHub**|Saisissez le nom de l'organisation GitHub.<br>Si vous n'en avez pas, créez en-une en consultant [comment créer une organisation dans Github](#how-to-create-organization-in-github).</br>|
|**Nom d'utilisateur GitHub**|Renseignez le nom d'utilisateur de votre compte GitHub.|
|**Personal Access Token**|Fournissez votre personal access token (PAT). Il est utilisé comme mot de passe alternatif pour authentifier votre compte GitHub.<br>Si vous n'en avez pas, créez un PAT GitHub [ici](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token).</br>|

### Comment créer une organisation dans GitHub
**Note** : Nous ne recommandons PAS d'utiliser l'organisation GitHub qui contient votre code source.

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github/github-gitops-latest.mp4" caption="GitHub" %}

1. Créez un nouveau compte dans GitHub (si vous n'en avez pas).
1. Dans le coin supérieur droit de votre page GitHub, cliquez sur votre photo de profil, puis sur `Paramètres`.
1. Dans la section `Accès`, cliquez sur `Organisations`.
1. `Dans la section `Organisations`, cliquez sur `Nouvelle organisation.
1. Choisissez un [plan](https://docs.github.com/en/get-started/learning-about-github/githubs-products) pour votre organisation. Vous avez également la possibilité de choisir de créer une `organisation gratuite`.
1. Sur la` `page` Configurer votre organisation,`
   * Renseignez le` nom du compte de l'organisation`, `l'email de contact`.
   * Sélectionnez l'option à laquelle votre organisation appartient.
   * Vérifiez votre compte et cliquez sur `Suivant`.
   * Votre` nom d'organisation GitHub `sera créé`.`
1. Allez dans votre profil et cliquez sur `Vos organisations` pour afficher toutes les organisations que vous avez créées.

Pour plus d'informations sur les plans disponibles pour votre équipe, consultez les [produits GitHub](https://docs.github.com/en/get-started/learning-about-github/githubs-products). Vous pouvez également consulter la page de documentation officielle de l'[organisation GitHub](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations) pour plus de détails.

**Note** :

* repo - Contrôle total des référentiels privés (possibilité d'accéder au statut de commit, au statut de déploiement et aux répertoires publics).
* admin:org - Contrôle total des organisations et des équipes (accès en lecture et en écriture).
* delete\_repo - Accorde l'accès à la suppression de repo dans les répertoires privés.
## GitLab
Si vous sélectionnez `GitLab` en tant que fournisseur git, veuillez renseigner les champs suivants pour configurer GitOps :

|Champs|Description|
| :-: | :-: |
|**Hôte Git**|Ce champ indique l'URL du fournisseur Git sélectionné. <br>A titre d'exemple : <br>https://gitlab.com/ pour GitLab.|
|**Groupe ID GitLab**|Saisissez l'ID du groupe GitLab.<br>Si vous n'en avez pas, créez-en un en utilisant [Groupe ID GitLab](#how-to-create-organization-in-gitlab).</br>|
|**Nom d'utilisateur GitLab**|Renseignez le nom d'utilisateur de votre compte GitLab.|
|**Personal Access Token**|Fournissez votre personal access token (PAT). Il est utilisé comme mot de passe alternatif pour authentifier votre compte GitLab.<br>Si vous n'en avez pas, créez un GitLab PAT [ici](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html).</br>|

### Comment créer une organisation dans GitLab
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab/gitops-gitlab-latest1.mp4" caption="GitHub" %}

1. Créez un nouveau compte sur GitLab (si vous n'en avez pas).
1. Vous pouvez créer un groupe en allant dans l'onglet 'Groupes' du tableau de bord GitLab puis en cliquant sur `Nouveau groupe`.
1. Sélectionnez` Créer un groupe`.
1. Saisissez le nom du groupe (obligatoire) et sélectionnez les descriptions facultatives, si nécessaire, puis cliquez sur `Créer groupe`.
1. Votre groupe sera créé et votre nom de groupe sera attribué avec un nouvel `ID groupe` (par exemple, 61512475).

**Note** :

* api - Confère un accès complet en lecture/écriture à l'API du projet concerné.
* write\_repository - Accorde l'accès en lecture/écriture (pull, push) au repository.
## Azure
Si vous sélectionnez `GitAzureLab` comme fournisseur git, veuillez renseigner les informations dans les champs suivants pour configurer GitOps :

|Champs|Description|
| :-: | :-: |
|**Url Organisation Azure DevOps**\*|Ce champ indique l'URL du fournisseur Git sélectionné. <br>A titre d'exemple :<br>https://dev.azure.com/ pour Azure.|
|**Nom du projet Azure DevOps**|Saisissez le nom du projet Azure DevOps.<br>Si vous n'en avez pas, créez en utilisant [Azure DevOps Project Name](#how-to-create-azure-devops-project-name).</br>|
|**Nom d'utilisateur Azure DevOps**\*|Renseignez le nom d'utilisateur de votre compte Azure DevOps.|
|**Jeton d'accès à Azure DevOps** \*|Renseignez votre jeton d'accès Azure DevOps. Il est utilisé comme mot de passe alternatif pour authentifier votre compte Azure DevOps.<br>Si vous n'en avez pas, créez un jeton d'accès Azure DevOps [ici](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page).</br>|

### Comment créer un nom de projet Azure DevOps
**Note** : Vous devez disposer d'une organisation avant de pouvoir créer un projet. Si vous n'avez pas encore créé d'organisation, créez-en une en suivant les instructions de la section[ S'inscrire, se connecter à Azure DevOps](https://learn.microsoft.com/en-us/azure/devops/user-guide/sign-up-invite-teammates?view=azure-devops), qui permet aussi de créer un projet. Ou consultez [Créer une organisation ou une collection de projets](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/create-organization?view=azure-devops).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-new-project.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-create-new-project.jpg)

1. Allez à Azure DevOps et naviguez jusqu'à Projets.
1. Sélectionnez votre organisation et cliquez sur `Nouveau projet`.
1. Dans la` `page` Créer un nouveau projet,`
   * Saisissez le `nom du projet` et sa description.
   * Sélectionnez l'option de visibilité (privée ou publique), le type de contrôle de la source initiale et le processus de l'élément de travail.
   * Cliquez sur `Créer`.
   * Azure DevOps affiche la page de bienvenue du projet avec le `nom du projet`.

Vous pouvez également vous référer à la page de la doc officielle [Azure DevOps Nom du projet](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page) pour plus de détails.

**Note** :

* code - Confère le droit de lecture du code source et des métadonnées sur les commits, les ensembles de modifications, les branches et autres artefacts de contrôle de version. [Plus d'informations sur les scopes dans Azure devops.](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes)
## BitBucket Cloud
Si vous sélectionnez `Bitbucket Cloud` comme fournisseur git, veuillez renseigner les champs suivants pour configurer GitOps :

|Champs|Description|
| :-: | :-: |
|**Hôte Bitbucket**|Ce champ indique l'URL du fournisseur Git sélectionné. <br>A titre d'exemple : <br>https://bitbucket.org/ pour Bitbucket.|
|**ID de l'espace de travail Bitbucket**|Saisissez l'ID de l'espace de travail Bitbucker.<br>Si vous n'en avez pas, créez en utilisant l'[Id de l'espace de travail Bitbucket](#how-to-create-bitbucket-workspace-id).</br>|
|**Clé de projet Bitbucket**|Saisissez la clé du projet Bitbucket.<br>Si vous n'en avez pas, créez-la en utilisant [Clé de projet Bitbucket](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/).</br><br>Note : Ce champ n'est pas obligatoire. Si le projet n'est pas fourni, le repository est automatiquement attribué au projet le plus ancien de l'espace de travail.</br>|
|**Nom d'utilisateur Nom d'utilisateur Bitbucket**\*|Renseignez le nom d'utilisateur de votre compte Bitbucket.|
|**Personal Access Token**|Fournissez votre personal access token (PAT). Il est utilisé comme mot de passe alternatif pour authentifier votre compte Bitbucket Cloud.<br>Si vous n'en avez pas, créez un PAT Bitbucket Cloud [ici](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/).</br>|

### Comment créer un ID d'espace de travail Bitbucket
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket/bitbucket-latest-gitops.mp4" caption="GitHub" %}

1. Créez un nouveau compte individuel sur Bitbucket (si vous n'en avez pas).
1. Sélectionnez votre profil et vos paramètres avatar dans le coin supérieur droit de la barre de navigation supérieure.
1. Sélectionnez `Tous les espaces de travail` dans le menu déroulant.
1. Sélectionnez `Créer un espace de travail` dans le coin supérieur droit de la page `Espaces de travail`.
1. Dans la` `page` Créer un espace de travail :`
* Saisissez un` nom d'espace de travail`.
* Saisissez un `Workspace ID`. Votre ID ne peut pas comporter d'espaces ou de caractères spéciaux, mais des chiffres et des lettres majuscules sont acceptés. Cet ID devient une partie de l'URL pour l'espace de travail et partout ailleurs où se trouve une étiquette identifiant l'équipe (API, groupes de permissions, OAuth, etc.).
* Cliquez sur `Créer`.
6. Votre `nom d'espace de travail` et votre` ID d'espace de travail` seront créés.

Vous pouvez également consulter la page de documentation officielle de l'[Id de l'espace de travail de Bitbucket](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/) pour plus de détails.

**Note** :

* repo - Gestion complète de l'accès aux référentiels (Lecture, Écriture, Admin, Suppression).

Cliquez sur **Sauvegarder** pour enregistrer les détails de votre configuration GitOps.

**Note** : Une coche verte apparaîtra sur le fournisseur GitOp actif.
