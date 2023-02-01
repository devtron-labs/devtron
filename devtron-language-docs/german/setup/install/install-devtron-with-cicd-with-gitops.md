# Devtron mit CI/CD zusammen mit GitOps installieren (Argo CD)
In diesem Abschnitt beschreiben wir die Schritte im Detail, wie Sie Devtron mit CI/CD installieren können, indem Sie GitOps während der Installation aktivieren.
## Bevor Sie anfangen
Installieren Sie [Helm](https://helm.sh/docs/intro/install/) falls Sie es noch nicht installiert haben.
## Devtron mit CI/CD zusammen mit GitOps installieren (Argo CD)
Führen Sie den folgenden Befehl aus, um die neueste Version von Devtron mit CI/CD zusammen mit dem Modul GitOps (Argo CD) zu installieren:
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set argo-cd.enabled=true
~~~

**Hinweis**: Wenn Sie den Blob-Speicher während der Installation konfigurieren möchten, lesen Sie bitte [Blob-Speicher während der Installation konfigurieren](#configure-blob-storage-duing-installation).
## Installieren Sie Multi-Architektur-Knoten (ARM und AMD)
Um Devtron auf Clustern mit Multi-Architektur-Knoten (ARM und AMD) zu installieren, fügen Sie dem Devtron-Installationsbefehl diese Option hinzu: `--set installer.arch=multi-arch`.

**Hinweis**:

* Um eine bestimmte Version von Devtron mit `vx.x.x` als [Release-Tag](https://github.com/devtron-labs/devtron/releases) zu installieren, fügen Sie dem Befehl den Zusatz hinzu: `--set installer.release="vX.X.X"`.
* Wenn Sie Devtron für `Production-Deployments` installieren möchten, beachten Sie bitte unsere empfohlenen Überschreibungen für die [Devtron-Installation](override-default-devtron-installation-configs.md).
## Blob Storage während der Installation konfigurieren
Die Konfiguration von Blob Storage in Ihrer Devtron-Umgebung ermöglicht es Ihnen, Build-Logs und Cache zu speichern.
Wenn Sie den Blob Storage nicht konfigurieren, dann:

- Sie können nach einer Stunde nicht mehr auf die Build- und Deployment-Logs zugreifen.
- Die Erstellungszeit für den Commit-Hash dauert länger, da der Cache nicht verfügbar ist.
- Artefaktberichte können in den Phasen vor und nach der Erstellung und Deployment nicht erstellt werden.

Wählen Sie eine der Optionen, um den Blob Storage zu konfigurieren:

{% tabs %}

{% tab title="MinIO Storage" %}

Führen Sie den folgenden Befehl aus, um Devtron zusammen mit MinIO für die Speicherung von Logs und Cache zu installieren.
~~~ bash
helm repo add devtron https://helm.devtron.ai 

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set installer.modules={cicd} \
--set minio.enabled=true \
--set argo-cd.enabled=true
~~~

**Hinweis**: Im Gegensatz zu globalen Cloud-Anbietern wie AWS S3 Bucket, Azure Blob Storage und Google Cloud Storage, kann MinIO auch lokal gehostet werden.

{% endtab %}

{% tab title="AWS S3 Bucket" %}

Beachten Sie die `AWS-spezifischen` Parameter auf der Seite [Storage für Logs und Cache](./installation-configuration.md#aws-specific).

Führen Sie den folgenden Befehl aus, um Devtron zusammen mit AWS S3-Buckets zum Speichern von Build-Logs und Cache zu installieren:

* Installation nach S3 IAM-Richtlinie.
> Hinweis: Bitte stellen Sie sicher, dass die S3-Berechtigungsrichtlinie für die IAM-Rolle den Knoten des Clusters zugeordnet ist, wenn Sie den folgenden Befehl verwenden.
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

* Installation unter Verwendung von access-key und secret-key für die AWS S3-Authentifizierung:
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

* Installation mit S3-kompatiblen Speichern:
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

Beachten Sie die `Azure-spezifischen` Parameter auf der Seite [Speicher für Logs und Cache](./installation-configuration.md#azure-specific).

Führen Sie den folgenden Befehl aus, um Devtron zusammen mit Azure Blob Storage zum Speichern von Build-Logs und Cache zu installieren:
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

Beachten Sie die `Google Cloud-spezifischen` Parameter auf der Seite [Speicher für Logs und Cache](./installation-configuration.md#google-cloud-storage-specific).

Führen Sie den folgenden Befehl aus, um Devtron zusammen mit Google Cloud Storage zum Speichern von Build-Logs und Cache zu installieren:
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
## Status der Devtron-Installation überprüfen
**Hinweis**: Die Installation dauert etwa 15 bis 20 Minuten, um alle Devtron-Microservices nacheinander hochzufahren.

Führen Sie den folgenden Befehl aus, um den Status der Installation zu überprüfen:
~~~ bash
kubectl -n devtroncd get installers installer-devtron \
-o jsonpath='{.status.sync.status}'
~~~

Der Befehl wird mit einer der folgenden Ausgabemeldungen ausgeführt, die den Status der Installation angeben:

|Status|Beschreibung|
| :- | :- |
|`Heruntergeladen`|Das Installationsprogramm hat alle Manifeste heruntergeladen, und die Installation wird durchgeführt.|
|`Angewandt`|Das Installationsprogramm hat alle Manifeste erfolgreich angewendet und die Installation ist abgeschlossen.|

## Installations-Logs prüfen
Führen Sie den folgenden Befehl aus, um die Logs des Installationsprogramms zu überprüfen:
~~~ bash
kubectl logs -f -l app=inception -n devtroncd
~~~
## Devtron-Dashboard
Führen Sie den folgenden Befehl aus, um die URL des Devtron-Dashboards zu erhalten:
~~~ bash
kubectl get svc -n devtroncd devtron-service \
-o jsonpath='{.status.loadBalancer.ingress}'
~~~

Sie erhalten eine Ausgabe ähnlich dem folgenden Beispiel:
~~~ bash
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

Verwenden Sie den Hostnamen` aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com `(Loadbalancer URL), um auf das Devtron-Dashboard zuzugreifen.

**Hinweis**: Wenn Sie keinen Hostnamen oder die Meldung „Dienst existiert nicht“ erhalten, bedeutet dies, dass Devtron noch installiert wird.
Bitte warten Sie, bis die Installation abgeschlossen ist.

**Hinweis**: Sie können auch einen `CNAME`-Eintrag verwenden, der Ihrer Domain/Subdomain entspricht, um auf die Loadbalancer-URL zu verweisen, damit der Zugriff über eine benutzerdefinierte Domain erfolgt.

|Host|Typ|Verweist auf|
| :- | :- | :- |
|devtron.yourdomain.com|CNAME|aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|

## Devtron-Admin-Zugangsdaten
### Für Devtron-Versionen v0.6.0 und später
**Benutzername**: `admin` <br>
**Passwort**: Führen Sie den folgenden Befehl aus, um das Admin-Passwort zu erhalten:
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d
~~~
### Für Devtron-Versionen vor v0.6.0
**Benutzername**: `admin` <br>
**Passwort**: Führen Sie den folgenden Befehl aus, um das Admin-Passwort zu erhalten:
~~~ bash
kubectl -n devtroncd get secret devtron-secret \
-o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
~~~

* Wenn Sie Devtron deinstallieren oder den Devtron-Helm-Installer bereinigen möchten, lesen Sie bitte unsere Anleitung zum [Deinstallieren von Devtron](https://docs.devtron.ai/install/uninstall-devtron).
* Bezüglich der Installation lesen Sie bitte auch den Abschnitt [FAQ](https://docs.devtron.ai/install/faq-on-installation).

**Hinweis**: Wenn Sie Fragen haben, lassen Sie es uns bitte auf unserem Discord-Channel wissen. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
