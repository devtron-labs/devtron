# Installer Devtron avec CI/CD en parallèle avec GitOps (Argo CD)
Dans cette section, nous décrivons en détail les étapes de l'installation de Devtron avec CI/CD en activant GitOps lors de l'installation.
## Avant de commencer
Installez [Helm](https://helm.sh/docs/intro/install/) si vous ne l'avez pas installé.
## Installer Devtron avec CI/CD en parallèle avec GitOps (Argo CD)
Exécutez la commande suivante pour installer la dernière version de Devtron avec CI/CD ainsi que le module GitOps (Argo CD) :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set argo-cd.enabled=true
~~~

**Note** : Si vous souhaitez configurer le stockage Blob pendant l'installation, reportez-vous à [configurer le stockage blob pendant l'installation](#configure-blob-storage-duing-installation).
## Installer des nœuds multi-architecture (ARM et AMD)
Pour installer Devtron sur des clusters avec des nœuds multi-architecture (ARM et AMD), ajoutez à la commande d'installation de Devtron `--set installer.arch=multi-arch`.

**Note** :

* Pour installer une version particulière de Devtron où `vx.x.x` est le [tag de la version](https://github.com/devtron-labs/devtron/releases), ajoutez à la commande `--set installer.release="vX.X.X"`.
* Si vous souhaitez installer Devtron pour des `déploiements de production`, veuillez vous référer aux remplacements recommandés pour l'[installation de Devtron](override-default-devtron-installation-configs.md).
## Configurer le stockage Blob après l'installation
La configuration du stockage Blob dans votre environnement Devtron vous permet de stocker les journaux de création et le cache.
Dans le cas où vous ne configurez pas le stockage Blob, alors :

- Vous ne pourrez pas accéder aux journaux de création et de déploiement après une heure.
- Le temps de création pour le commit hash est plus long car le cache n'est pas disponible.
- Les rapports d'artefact ne peuvent pas être générés dans les étapes de pré/post création et de déploiement.

Choisissez l'une des options pour configurer le stockage blob :

{% tabs %}

{% tab title="MinIO Storage" %}

Exécutez la commande suivante pour installer Devtron avec MinIO pour le stockage des journaux et du cache.
~~~ bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true \
--set argo-cd.enabled=true
~~~

**Note** : Contrairement aux fournisseurs de cloud mondiaux tels que AWS S3 Bucket, Azure Blob Storage et Google Cloud Storage, MinIO peut également être hébergé localement.

{% endtab %}

{% tab title="AWS S3 Bucket" %}

Reportez-vous aux paramètres `spécifiques à AWS` sur la page [Stockage pour les journaux et le cache](./installation-configuration.md#aws-specific).

Exécutez la commande suivante pour installer Devtron ainsi que les buckets AWS S3 pour stocker les journaux de création et le cache :

* Installer en utilisant la politique S3 IAM.
> Note : Veuillez vous assurer de l'existence de la politique de permission S3 pour le rôle IAM attaché aux nœuds du cluster si vous utilisez la commande ci-dessous.
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set argo-cd.enabled=true
~~~

* Installer en utilisant access-key et secret-key pour l'authentification AWS S3 :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key> \
--set argo-cd.enabled=true
~~~

* Installer en utilisant des stockages compatibles S3 :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=S3 \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-1 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-1 \
--set secrets.BLOB_STORAGE_S3_ACCESS_KEY=<access-key> \
--set secrets.BLOB_STORAGE_S3_SECRET_KEY=<secret-key> \
--set configs.BLOB_STORAGE_S3_ENDPOINT=<endpoint> \
--set argo-cd.enabled=true
~~~

{% endtab %}

{% tab title="Azure Blob Storage" %}

Reportez-vous aux paramètres `spécifiques à Azure` sur la page [Stockage pour les journaux et le cache](./installation-configuration.md#azure-specific).

Exécutez la commande suivante pour installer Devtron avec Azure Blob Storage pour stocker les journaux de création et le cache :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set secrets.AZURE_ACCOUNT_KEY=xxxxxxxxxx \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container \
--set argo-cd.enabled=true
~~~

{% endtab %}

{% tab title="Google Cloud Storage" %}

Reportez-vous aux paramètres `spécifiques à Google Cloud` sur la page [Stockage pour les journaux et le cache](./installation-configuration.md#google-cloud-storage-specific).

Exécutez la commande suivante pour installer Devtron avec Google Cloud Storage pour stocker les journaux de création et le cache :
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set configs.BLOB_STORAGE_PROVIDER=GCP \
--set secrets.BLOB_STORAGE_GCP_CREDENTIALS_JSON=eyJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsInByb2plY3RfaWQiOiAiPHlvdXItcHJvamVjdC1pZD4iLCJwcml2YXRlX2tleV9pZCI6ICI8eW91ci1wcml2YXRlLWtleS1pZD4iLCJwcml2YXRlX2tleSI6ICI8eW91ci1wcml2YXRlLWtleT4iLCJjbGllbnRfZW1haWwiOiAiPHlvdXItY2xpZW50LWVtYWlsPiIsImNsaWVudF9pZCI6ICI8eW91ci1jbGllbnQtaWQ+IiwiYXV0aF91cmkiOiAiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLCJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLCJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vb2F1dGgyL3YxL2NlcnRzIiwiY2xpZW50X3g1MDlfY2VydF91cmwiOiAiPHlvdXItY2xpZW50LWNlcnQtdXJsPiJ9Cg== \
--set configs.DEFAULT_CACHE_BUCKET=cache-bucket \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=log-bucket \
--set argo-cd.enabled=true
~~~

{% endtab %}
{% endtabs %}
## Vérifier le statut de l'installation de Devtron
**Note** : L'installation prend environ 15 à 20 minutes pour faire tourner tous les microservices Devtron un par un.

Exécutez la commande suivante pour vérifier le statut de l'installation :
~~~ bash
kubectl -n devtroncd get installers installer-devtron \
-o jsonpath='{.status.sync.status}'
~~~

La commande s'exécute avec l'un des messages de sortie suivants, indiquant le statut de l'installation :

|Statut|Description|
| :- | :- |
|`Téléchargé`|Le programme d'installation a téléchargé tous les manifestes, et l'installation est en cours.|
|`Appliqué`|Le programme d'installation a appliqué avec succès tous les manifestes, et l'installation est terminée.|

## Vérifiez les journaux d'installation
Exécutez la commande suivante pour vérifier les journaux d'installation :
~~~ bash
kubectl logs -f -l app=inception -n devtroncd
~~~
## Tableau de bord Devtron
Exécutez la commande suivante pour obtenir l'URL du tableau de bord Devtron :
~~~ bash
kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
~~~

Vous obtiendrez un résultat similaire à l'exemple ci-dessous :
~~~ bash
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

Utilisez le nom d'hôte `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` (Loadbalancer URL) pour accéder au tableau de bord Devtron.

**Note** : Si vous n'obtenez pas de nom d'hôte ou si vous recevez un message indiquant " le service n'existe pas ", cela signifie que Devtron est toujours en cours d'installation.
Veuillez attendre la fin de l'installation.

**Note** : Vous pouvez également utiliser une entrée `CNAME` correspondant à votre domaine/sous-domaine pour pointer vers l'URL du Loadbalancer afin d'accéder à un domaine personnalisé.

|Hôte|Type|Pointe vers|
| :- | :- | :- |
|devtron.yourdomain.com|CNAME|aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|

## Identifiants Admin Devtron
### Pour la version Devtron v0.6.0 et ultérieure
**Nom d'utilisateur** : `admin` <br>
**Mot de passe** : Exécutez la commande suivante pour obtenir le mot de passe admin :
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
~~~
### Pour les versions de Devtron antérieures à la v0.6.0
**Nom d'utilisateur** : `admin` <br>
**Mot de passe** : Exécutez la commande suivante pour obtenir le mot de passe admin :
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
~~~

* Si vous souhaitez désinstaller Devtron ou nettoyer le programme d'installation Helm Devtron, consultez notre rubrique [Désinstaller Devtron](https://docs.devtron.ai/install/uninstall-devtron).
* En ce qui concerne l'installation, veuillez également consulter la section [FAQ](https://docs.devtron.ai/install/faq-on-installation).

**Note** : Si vous avez des questions, veuillez nous en faire part sur notre canal discord. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
