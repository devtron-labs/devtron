\# Container Registries

Container-Registries dienen zum Speichern von Images , die von der CI-Pipeline erstellt wurden. Sie können die Container-Registry mit jedem Container-Registry -Anbieter Ihrer Wahl konfigurieren. Auf diese Weise können Sie Ihre Container-Images mittels einfach bedienbarer Benutzeroberfläche (User Interface) aufbauen, einsetzen und verwalten.

Beim Konfigurieren einer Anwendung können Sie die spezifische Container Registry sowie das Repository in der App-Konfiguration auswählen > [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration).


\## Add Container Registry:

Um eine Container Registry hinzuzufügen, gehen Sie bitte zu „Container Registry‟ im Abschnitt „Global Configurations‟. Klicken Sie auf \*\*Add Container Registry\*\* (Container Registry hinzufügen).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/add-container-registry.jpg)

Tragen Sie die jeweiligen Informationen in die folgenden Felder ein, um die Container Registry hinzuzufügen.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Tragen Sie in Ihre Registry einen Namen ein. Dieser Name wird Ihnen in der Build-Konfiguration [Build Configuration](https://docs.devtron.ai/usage/applications/creating-application/docker-build-configuration) in der Dropdown-Liste angezeigt. |

| \*\*Registry-Typ\*\* | Wählen Sie den Registry-Typ in der Dropdown-Liste:<br><ul><li>[ECR](#registry-type-ecr)</li></ul><ul><li>[Docker](#registry-type-docker)</li></ul><ul><li>[Azure](#registry-type-azure)</li></ul><ul><li>[Artifact Registry (GCP)](#registry-type-artifact-registry-gcp)</li></ul><ul><li>[GCR](#registry-type-google-container-registry-gcr)</li></ul><ul><li>[Quay](#registry-type-quay)</li></ul><ul><li>[Other](#registry-type-other)</li></ul>Hinweis: Je nach \*\*Registry-Typ\*\* unterscheiden sich die Eingabefelder für die Anmeldeinformationen. |

| \*\*Registry URL\*\* | Tragen Sie die URL Ihrer Registry ein. |

| \*\*Als Standard-Registry festlegen\*\* | Dieses Feld aktivieren, um es als Standard-Registry-Hub für Ihre Images zu bestimmen. |



\### Registry Type: ECR

Amazon ECR ist ein von AWS verwalteter Container-Image-Registry-Service.

Das ECR stellt unter Verwendung von AWS Identity and Access Management (IAM) ressourcenbasierte Berechtigungen für private Repositories zur Verfügung. ECR gestattet neben Key-basierten auch Rollen-basierte Authentifizierungen.

Erstellen Sie zuerst einen IAM-Benutzer [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html) und ergänzen Sie die ECR-Richtlinie gemäß Authentifizierungstyp.

Tragen Sie die folgenden Informationen ein, falls Sie den Registry-Typ „ECR‟ wählen.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*ECR\*\*. |

| \*\*Registry-URL\*\* | Dies ist die URL Ihrer privaten Registrierung in AWS.<br>Beispielsweise das URL-Format: `https://xxxxxxxxxxxx.dkr.ecr.<region>.amazonaws.com`. `xxxxxxxxxxxx` ist Ihre 12-stellige AWS-Konto-ID.</br> |

| \*\*Authentifizierungstyp\*\* | Wählen Sie einen der Authentifizierungstypen:<ul><li>\*\*EC2 IAM-Rolle\*\*: Authentifizieren Sie sich mit der Workernode-IAM-Rolle und fügen Sie die ECR-Richtlinie (AmazonEC2ContainerRegistryFullAccess) der Cluster-Workernode-IAM-Rolle Ihres Kubernetes-Clusters hinzu.</li></ul><ul><li>\*\*Benutzer-Authentifizierung\*\*: Dies ist eine sclüsselbasierte Authentifizierung. Die ECR-Richtlinie (AmazonEC2ContainerRegistryFullAccess) ist dem IAM-Benutzer hinzuzufügen [IAM user](https://docs.aws.amazon.com/AmazonECR/latest/userguide/get-set-up-for-amazon-ecr.html).<ul><li>`Access key ID`: AWS-Zugangsschlüssel</li></ul><ul><li>`Secret access key`: Ihre geheime AWS-Zugangsschlüssel-ID</li></ul> |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um `ECR` als Standard-Registry-Hub für Ihre Images festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).


\### Registry Type: Docker

Tragen Sie bei der Auswahl des Registry-Typs „Docker‟ die folgenden Informationen an.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*Docker\*\*. |

| \*\*Registry-URL\*\* | Dies ist die URL Ihrer privaten Registry in Docker. Z. B. `docker.io` |

| \*\*Benutzername\*\* | Tragen Sie den Benutzernamen des Docker-Hub-Kontos ein, das Sie für die Erstellung Ihrer Registrierung benutzt haben. |

| \*\*Password/Token\*\* | Tragen Sie das entsprechende Password/[Token](https://docs.docker.com/docker-hub/access-tokens/) für Ihr Docker-Hub-Konto ein. Aus Sicherheitsgründen wird empfohlen, „Token‟ zu verwenden. |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um „Docker‟ für Ihre Images als Standard-Registry-Hub festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).

\### Registry Type: Azure

Für den Registry-Typ „Azure‟ kann die Authentifizierungsmethode des Service Principal (Dienstprinzipal) zur Authentifizierung mit Benutzername und Kennwort verwendet werden Folgen Sie bitte dem Link [link](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal), um den Benutzernamen und das Passwort für diese Registry zu erhalten.

Tragen Sie die folgenden Informationen ein, falls Sie als Registry-Typ „Azure‟ gewählt haben.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*Azure\*\*. |

| \*\*Registry-URL/Login Server\*\* | Dies ist die URL Ihrer privaten Registry in Azure. Z. B. `xxx.azurecr.io` |

| \*\*Benutzername/Registry-Name\*\* | Tragen Sie den Benutzernamen Ihrer Azure-Container-Registry ein. |

| \*\*Passwort\*\* | Tragen Sie das Passwort Ihrer Azure-Container-Registry ein. |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um „Azure‟ als Standard-Registry-Hub für Ihre Images festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).


\### Registry Type: Artifact Registry (GCP)

Die JSON-Schlüsseldatei-Authentifizierungsmethode kann zur Authentifizierung mit einer JSON-Datei für den Benutzernamen und den Service Account (das Dienstkonto) verwendet werden. Bitte folgen Sie dem Link [link](https://cloud.google.com/artifact-registry/docs/docker/authentication#json-key), um für diese Registry die JSON-Datei mit dem Benutzernamen und dem Dienstkonto zu erhalten.

\*\*Hinweis\*\*: Bitte löschen Sie zuvor alle Leerzeichen aus dem json-Schlüssel und fügen Sie ihn in einfache Anführungszeichen ein, wenn Sie ihn in das Feld „Service Account JSON File‟ eintragen.


Tragen Sie die folgenden Informationen ein, falls Sie den Registry-Typ „Artifact Registry (GCP)‟ auswählen.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*Artifact Registry (GCP)\*\*. |

| \*\*Registry-URL\*\* | Dies ist die URL Ihrer privaten Registry in Artifact Registry (GCP). Z. B. „region-docker.pkg.dev‟ |

| \*\*Benutzername\*\* | Tragen Sie den Benutzernamen des Artifact Registry (GCP) Kontos ein. |

| \*\*Service Account JSON File\*\* | Stellen Sie die Service Account JSON-Datei von Artifact Registry (GCP) bereit. |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um „Artifact Registry (GCP)‟ als Standard-Registry-Hub für Ihre Images festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).

\### Registry Type: Google Container Registry (GCR)

Die JSON-Schlüsseldatei-Authentifizierungsmethode kann zur Authentifizierung mit einer JSON-Datei für den Benutzernamen und den Service Account (das Dienstkonto) verwendet werden. Bitte folgen Sie dem Link [link](https://cloud.google.com/container-registry/docs/advanced-authentication#json-key), um den Benutzernamen und die JSON-Datei des Dienstkontos für diese Registry zu erhalten. Bitte löschen Sie alle Leerzeichen aus dem json-Schlüssel und schließen Sie ihn in einfache Anführungszeichen ein, wenn Sie ihn in das Feld „Service Account JSON File‟ eintragen.

Tragen Sie die folgenden Informationen ein, wenn Sie den Registry-Typ „GCR‟ auswählen

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*GCR\*\*. |

| \*\*Registry URL\*\* | Dies ist die URL Ihrer privaten Registry in GCR. Z. B. `gcr.io` |

| \*\*Benutzername\*\* | Tragen Sie den Benutzernamen Ihres GCR-Kontos ein. |

| \*\*Dienstkonto JSON File\*\* | Stellen Sie den Service Account JSON File von GCR bereit. |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um „GCR‟ als Standard-Registry für Ihre Images festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).

\### Registry Type: Quay

Tragen Sie die folgenden Informationen ein, wenn Sie als Registry-Typ „Quay‟ wählen.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*Quay\*\*. |

| \*\*Registry-URL\*\* | Dies ist die URL Ihrer privaten Registry in Quay. Z. B. `quay.io` |

| \*\*Benutzername\*\* | Tragen Sie den Benutzernamen des Quay-Kontos ein. |

| \*\*Passwort/Token\*\* | Stellen Sie das Passwort für Ihr Quay-Konto bereit. |

| \*\*Als Standard-Registry festlegen\*\* | Aktivieren Sie dieses Feld, um Quay als Standard-Registry-Hub für Ihre Images festzulegen. |

Klicken Sie auf \*\*Save\*\* (Speichern).


\### Registry Type: Other


Tragen Sie die folgenden Informationen ein, falls Sie den Registry-Typ „Anderer‟ wählen.

| Felder | Beschreibung |

\| --- | --- |

| \*\*Name\*\* | Benutzerdefinierter Name für die Registry in Devtron. |

| \*\*Registry-Typ\*\* | Wählen Sie \*\*Anderer\*\*. |

| \*\*Registry-URL\*\* | Dies ist die URL Ihrer privaten Registry. |

| \*\*Benutzername\*\* | Tragen Sie den Benutzernamen Ihres Kontos ein, in dem Sie Ihre Registry erstellt haben. |

| \*\*Passwort/Token\*\* | Tragen Sie das Passwort/Token ein, das dem Benutzernamen Ihrer Registry entspricht. |

| \*\*Als Standard-Registry festlegen\*\* | Dieses Feld aktivieren, um es als Standard-Registry-Hub für Ihre Images zu bestimmen. |

Klicken Sie auf \*\*Save\*\* (Speichern).

\#### Advance Registry URL Connection Options:

* Falls Sie die Option „Allow Only Secure Connection‟ (Nur sichere Verbindungen zulassen) aktivieren, gestattet diese Registry ausschließlich sichere Verbindungen.
* Falls Sie die Option „Allow Secure Connection With CA Certificate‟ (Sichere Verbindung mit CA-Zertifikat zulassen) aktivieren, müssen Sie ein privates CA-Zertifikat (ca.crt) hochladen/ bereitstellen.
* Falls die Container-Registry unsicher ist (z. B. : Das SSL-Zertifikat ist abgelaufen), aktivieren Sie die Option „Allow Insecure Connection‟.

\*\*Hinweis\*\*: Sie können jede Registry verwenden, die mit `docker login -u <username> -p <password> <registry-url>` authentifiziert werden kann. Allerdings können diese Registries einen sichereren Weg zur Authentifizierung bieten, den wir später unterstützen werden.


\## Pull an Image from a Private Registry

Sie können einen Pod erstellen, der ein „Secret‟ verwendet, um ein Abbild aus einer privaten Container-Registry zu ziehen. Sie können jede private Container-Registry Ihrer Wahl verwenden. Zum Beispiel: [Docker Hub](https://www.docker.com/products/docker-hub).

Super-Admin-Benutzer können entscheiden, ob sie die Anmeldeinformationen für die Registry automatisch einspeisen oder ein Secret einsetzen möchten, um ein Image für die Bereitstellung in Umgebungen auf bestimmten Clustern zu ziehen.

Klicken Sie auf \*\*Manage\*\* (Verwalten), um den Zugriff auf die Anmeldeinformationen für die Registry zu verwalten.

Es sind zwei Optionen verfügbar, um den Zugriff auf die Anmeldeinformationen zu verwalten:

| Felder | Beschreibung |

\| --- | --- |

| \*\*Speisen Sie keine Anmeldeinformationen in Cluster ein\*\* | Wählen Sie die Cluster aus, für die Sie keine Anmeldedaten eintragen möchten. |

| \*\*Automatische Eingabe von Anmeldeinformationen in Clustern\*\* | Wählen sie die Cluster aus, für die Sie Anmeldedaten eintragen möchten. |

Sie können eine der beiden Optionen für die Definition von Berechtigungsnachweisen wählen:

* [User Registry Credentials](#use-registry-credentials)
* [Specify Image Pull Secret](#specify-image-pull-secret)

\### Use Registry Credentials

Falls sie \*\*Use Registry Credentials\*\* (Registry-Daten verwenden) wählen, werden die Cluster automatisch mit den Registry-Daten Ihres Registry-Typs bestückt. Falls Sie beispielsweise „Docker‟ als Registry-Typ wählen, werden die Cluster automatisch mit dem „Benutzernamen‟ und dem „Passwort/Token‟ für Ihr Docker-Hub-Konto gefüllt.

Klicken Sie auf \*\*Save\*\* (Speichern).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/use-registry-credentials.jpg)


\### Specify Image Pull Secret

Sie können ein Secret erstellen, indem Sie Ihre Zugangsdaten in der Befehlszeile eintragen.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/container-registries/specify-image-pull-secret-latest.png)

Erstellen Sie dieses Secret und bezeichnen Sie es mit „regcred‟.

\```bash

kubectl create -n <namespace> secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>

\```

wo:

* <namespace> ist Ihr virtueller Cluster. Z. B. devtron-demo
* <your-registry-server> ist der FQDN Ihrer privaten Docker-Registry. Verwenden Sie https://index.docker.io/v1/ für DockerHub.
* <your-name> ist Ihr Docker-Benutzername.
* <your-pword> ist Ihr Docker-Passwort.
* <your-email> ist Ihre Docker-E-Mail.

Sie haben Ihre Docker-Anmeldeinformationen im Cluster erfolgreich als Secret mit dem Namen „regcred‟ festgelegt.

\*\*Hinweis\*\*: Der Eintrag von Secrets in der Kommandozeile kann dazu führen, dass diese ungeschützt in der Shell-Historie gespeichert werden und während der Ausführung von kubectl auch für andere Benutzer auf Ihrem PC sichtbar sind.

Tragen Sie die Bezeichnung „Secret‟ in das Feld ein und klicken Sie „Save‟ (Speichern).






















