\# Registres des conteneurs

Les registres de conteneurs sont utilisés pour stocker les images construites par le pipeline CI. Vous pouvez configurer le registre de conteneurs en utilisant le fournisseur de registre de conteneurs de votre choix. Il vous permet de construire, déployer et gérer vos images de conteneur avec une UI conviviale.

Lors de la configuration d'une application, vous pouvez choisir le registre de conteneurs et le référentiel spécifiques dans la section de Configuration de l'app > [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration).


\## Ajouter un registre de conteneurs :

Pour ajouter le registre des conteneurs, allez dans la section `Registre des conteneurs` de `Configurations générales`. Cliquez sur \*\*Ajouter un registre de conteneurs\*\*.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry.jpg)

Complétez les informations dans les champs suivants pour ajouter le registre des conteneurs.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Attribuez un nom à votre registre, ce nom sera affiché dans [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) dans la liste déroulante. |

| \*\*Type de registre\*\* | Sélectionnez le type de registre dans la liste déroulante :<br><ul><li>[ECR](#registry-type-ecr)</li></ul><ul><li>[Docker](#registry-type-docker)</li></ul><ul><li>[Azure](#registry-type-azure)</li></ul><ul><li>[Artifact Registry (GCP)](#registry-type-artifact-registry-gcp)</li></ul><ul><li>[GCR](#registry-type-google-container-registry-gcr)</li></ul><ul><li>[Quay](#registry-type-quay)</li></ul><ul><li>[Autre](#registry-type-other)</li></ul>`Note` : Pour chaque \*\*Type de registre\*\*, les champs de saisie des informations d'identification sont différents. |

| \*\*URL de registre\*\* | Indiquez l'URL de votre registre. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir comme hub de registre par défaut pour vos images. |



\### Type de registre : ECR

Amazon ECR est un service de registre d'images de conteneurs géré par AWS.

L'ECR fournit des autorisations basées sur les ressources aux référentiels privés en utilisant AWS Identity and Access Management (IAM). ECR permet les authentifications basées sur les clés et sur les rôles.

Avant de commencer, créez un [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html) et attachez la politique ECR en fonction du type d'authentification.

Renseignez les informations ci-dessous si vous sélectionnez le type de registre `ECR`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*ECR\*\*. |

| \*\*URL de registre\*\* | Il s'agit de l'URL de votre registre privé dans AWS.<br>Par exemple, le format de l'URL est : `https://xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com`. `xxxxxxxxxxxx` est votre ID de compte AWS à 12 chiffres.</br> |

| \*\*Type d'authentification\*\* | Sélectionnez l'un des types d'authentification : <ul><li>\*\*Rôle IAM EC2\*\* : Authentifiez-vous avec le rôle IAM de workernode et attachez la politique ECR (AmazonEC2ContainerRegistryFullAccess) au rôle IAM de workernode de votre cluster Kubernetes.</li></ul><ul><li>\*\*Authentification de l'utilisateur\*\*: Il s'agit d'une authentification basée sur une clé et d'attacher la politique ECR (AmazonEC2ContainerRegistryFullAccess) à [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html).<ul><li>`Access key ID` : Votre clé d'accès AWS</li></ul><ul><li>`Secret access key` : Votre ID de clé d'accès secrète AWS</li></ul> |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `ECR` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.


\### Type de registre : Docker

Renseignez les informations ci-dessous si vous sélectionnez le type de registre `Docker`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*Docker\*\*. |

| \*\*L'URL du registre\*\* | Il s'agit de l'URL de votre registre privé dans Docker. Par exemple, `docker.io` |

| \*\*Nom d'utilisateur\*\* | Fournissez le nom d'utilisateur du compte docker hub que vous avez utilisé pour créer votre registre. |

| \*\*Mot de passe/Jeton\*\* | Renseignez le mot de passe/[Token](https://docs.docker.com/docker-hub/access-tokens/) correspondant à votre compte docker hub. Il est recommandé d'utiliser `Jeton` pour des raisons de sécurité. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `Docker` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.

\### Type de registre : Azure

Pour le type de registre : Azure, la méthode d'authentification du principal du service peut être utilisée pour s'authentifier avec un nom d'utilisateur et un mot de passe. Consultez [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal) pour obtenir le nom d'utilisateur et le mot de passe de ce registre.

Renseignez les informations ci-dessous si vous sélectionnez le type de registre `Azure`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*Azure\*\*. |

| \*\* URL du registre/Serveur de connexion\*\* | Il s'agit de l'URL de votre registre privé dans Azure. Par exemple, `xxx.azurecr.io`

| Nom d'utilisateur/Nom de registre\*\* | Renseignez le nom d'utilisateur du registre de conteneurs Azure. |

| \*\*Mot de passe\*\* | Renseignez le mot de passe du registre des conteneurs Azure. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `Azure` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.


\#### Type de registre : Artifact Registry (GCP)

La méthode d'authentification par fichier clé JSON peut être utilisée pour s'authentifier avec le nom d'utilisateur et le fichier JSON du compte de service. Veuillez consulter [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key) pour obtenir le nom d'utilisateur et le fichier JSON du compte de service pour ce registre.

\*\*Note\*\* : Veuillez supprimer tous les espaces blancs de la clé json et la mettre entre guillemets lorsqu'elle est placée dans le champ `Service Account JSON File`.


Renseignez les informations ci-dessous si vous sélectionnez le type de registre comme `Artifact Registry (GCP)`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*Artifact Registry (GCP)\*\*. |

| \*\* URL du registre\*\* | Il s'agit de l'URL de votre registre privé dans l'Artifact Registry (GCP). Par exemple, `region-docker.pkg.dev` |

| \*\*Nom d'utilisateur\*\* | Renseignez le nom d'utilisateur du compte du Artifact Registry (GCP). |

| \*\*Fichier JSON du compte de service\*\* | Renseignez le fichier JSON du Compte de service du Artifact Registry (GCP). |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `Artifact Registry (GCP)` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.

\### Type de registre : Google Container Registry (GCR)

La méthode d'authentification par fichier clé JSON peut être utilisée pour s'authentifier avec le nom d'utilisateur et le fichier JSON du compte de service. Veuillez consulter [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key) pour obtenir le nom d'utilisateur et le fichier JSON du compte de service pour ce registre. Veuillez supprimer tous les espaces blancs de la clé json et la mettre entre guillemets lorsqu'elle est placée dans le champ `Service Account JSON File`.

Renseignez les informations ci-dessous si vous sélectionnez le type de registre `GCR`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*GCR\*\*. |

| \*\*URL du registre\*\* | Il s'agit de l'URL de votre registre privé dans GCR. Par exemple, `gcr.io` |

| \*\*Nom d'utilisateur\*\* | Renseignez le nom d'utilisateur de votre compte GCR. |

| \*\*Fichier JSON du compte de service\*\* | Renseignez le fichier JSON du compte de service du GCR. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `GCR` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.

\### Type de registre : Quay

Renseignez les informations ci-dessous si vous sélectionnez le type de registre `Quay`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*Quay\*\*. |

| \*\*URL du registre\*\* | Il s'agit de l'URL de votre registre privé dans Quay. Par exemple, `quay.io` |

| \*\*Nom d'utilisateur :\*\* | Indiquez le nom d'utilisateur du compte Quay. |

| \*\*Mot de passe/Jeton\*\* | Indiquez le mot de passe de votre compte Quay. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir `Quay` comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.


\### Type de registre : Autre


Renseignez les informations ci-dessous si vous sélectionnez le type de registre `Autre`.

| Champs | Description |

\| --- | --- |

| \*\*Nom\*\* | Nom défini par l'utilisateur pour le registre dans Devtron. |

| \*\*Type de registre\*\* | Sélectionnez \*\*Autre\*\*. |

| \*\* URL du registre\*\* | Il s'agit de l'URL de votre registre privé. |

| \*\*Nom d'utilisateur\*\* | Renseignez le nom d'utilisateur de votre compte où vous avez créé votre registre. |

| \*\*Mot de passe/Jeton\*\* | Renseignez le mot de passe/Jeton correspondant au nom d'utilisateur de votre registre. |

| \*\*Définir comme registre par défaut\*\* | Activez ce champ pour définir comme hub de registre par défaut pour vos images. |

Cliquez sur \*\*Sauvegarder\*\*.

\#### Options avancées de connexion à l'URL du registre :

* Si vous activez l'option `Autoriser uniquement les connexions sécurisées`, alors ce registre n'autorise que les connexions sécurisées.
* Si vous activez l'option `Connexion sécurisée faible avec le certificat CA`, vous devez ensuite télécharger/fournir un certificat CA privé (ca.crt).
* Si le registre du conteneur n'est pas sécurisé (par exemple : le certificat SSL est expiré), vous devez activer l'option`Autoriser les connexions non sécurisées`.

\*\*Note\*\* : Vous pouvez utiliser n'importe quel registre qui peut être authentifié en utilisant `docker login -u <username> -p <password> <registry-url>`. Cependant, ces registres peuvent fournir une méthode d'authentification plus sécurisée, que nous prendrons en charge plus tard.


\## Extraire une image d'un registre privé

Vous pouvez créer un Pod qui utilise un `Secret` pour extraire une image d'un registre de conteneurs privé. Vous pouvez utiliser le registre de conteneurs privés de votre choix. À titre d'exemple : [Docker Hub](https://www.docker.com/products/docker-hub).

Les utilisateurs super admin peuvent décider s'ils veulent injecter automatiquement les informations d'identification du registre ou utiliser un secret pour extraire une image à déployer dans des environnements sur des clusters spécifiques.

Pour gérer l'accès aux informations d'identification du registre, cliquez sur \*\*Gérer\*\*.

Il existe deux options pour gérer l'accès aux informations d'identification du registre :

| Champs | Description |

\| --- | --- |

| \*\*Ne pas injecter les informations d'identification aux clusters\*\* | Sélectionnez les clusters pour lesquels vous ne voulez pas injecter les informations d'identification. |

| \*\*Auto-injection d'informations d'identification aux clusters\*\* | Sélectionnez les clusters pour lesquels vous souhaitez injecter des informations d'identification. |

Vous pouvez choisir l'une des deux options pour définir les informations d'identification :

* [Identifiants du registre des utilisateurs](#use-registry-credentials)
* [Spécifier l'Image Pull Secret](#specify-image-pull-secret)

\### Utiliser les identifiants du registre

Si vous sélectionnez \*\*Utiliser les identifiants de registre\*\*, les clusters seront injectés automatiquement avec les identifiants de registre de votre type de registre. A titre d'exemple, si vous sélectionnez `Docker` comme Type de Registre, alors les clusters seront auto-injectés avec le `username` et le `password/token` que vous utilisez sur le compte Docker Hub.

Cliquez sur \*\*Sauvegarder\*\*.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials.jpg)


\### Spécifiez l'Image Pull Secret

Vous pouvez créer un Secret en fournissant des informations d'identification sur la ligne de commande.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-latest.png)

Créez ce Secret, en le nommant `regcred` :

\```bash

kubectl create -n <namespace> secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>

\```

où :

* <namespace> est votre cluster virtuel. Par exemple, devtron-demo
* <your-registry-server> est votre FQDN de registre Docker privé. Utilisez https://index.docker.io/v1/ pour DockerHub.
* <your-name> est votre nom d'utilisateur Docker.
* <your-pword> est votre mot de passe Docker.
* <your-email> est votre email Docker.

Vous avez réussi à définir vos informations d'identification Docker dans le cluster en tant que Secret appelé `regcred`.

\*\*Note\*\* : Saisir des secrets sur la ligne de commande peut les stocker dans l'historique de votre shell sans protection, et ces secrets peuvent également être visibles par d'autres utilisateurs sur votre PC pendant la durée d'exécution de kubectl.

Saisissez le nom du `Secret` dans le champ et cliquez sur \*\*Sauvegarder\*\*.






















