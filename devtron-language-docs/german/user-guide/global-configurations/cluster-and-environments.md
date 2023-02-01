# Cluster und Umgebungen
Sie können Ihre vorhandenen Kubernetes-Cluster und -Umgebungen im Bereich `Clusters and Environments` hinzufügen. Zum Hinzufügen eines Clusters müssen Sie über einen [Super-Admin](https://docs.devtron.ai/global-configurations/authorization/user-access#assign-super-admin-permissions)-Zugang verfügen.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster-and-environments.png)
## Cluster hinzufügen:
Zum Hinzufügen eines Clusters gehen Sie bitte zu `Clusters & Environments` im Abschnitt `Global Configurations`. Bitte **Add cluster** anklicken.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/add-clusters.png)

Füllen Sie die folgenden Felder entsprechend aus, um Ihre Kubernetes-Cluster hinzuzufügen:

|Feld|Beschreibung|
| :- | :- |
|`मुं`|Tragen Sie die Bezeichnung Ihres Clusters ein.|
|`Server URL`|Server URL eines Clusters.<br>Hinweis: Wir empfehlen die Verwendung einer [selbst gehosteten URL](#benefits-of-self-hosted-url) anstelle einer in der Cloud gehosteten URL.</br>|
|`Inhaber-Token`|Inhaber-Token eines Clusters.|

### Cluster-Anmeldedaten abrufen
> **Voraussetzungen:** `kubectl` und `jq` müssen auf der Bastion vorinstalliert sein.

**Hinweis**: Wir empfehlen die Verwendung einer selbst gehosteten URL anstelle einer in der Cloud. Informieren Sie sich über die Vorteile einer [selbst gehosteten URL](#benefits-of-self-hosted-url).

Sie können **`Server URL`** und **`Bearer Token`** abrufen, indem Sie abhängig vom Cluster-Provider den folgenden Befehl ausführen:

{% tabs %}
{% tab title="k8s Cluster Providers" %}
Falls Sie EKS, AKS, GKE, Kops oder von Digital Ocean verwaltete Kubernetes verwenden, führen Sie den folgenden Befehl aus, um Server-URL und Bearer-Token zu generieren:
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh \
&& bash kubernetes_export_sa.sh cd-user devtroncd \
https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% tab title="Microk8s Cluster" %}
Falls sie einen **`microk8s cluster`** verwenden, führen Sie den folgenden Befehl aus, um Server URL und Bearer-Token zu generieren:
~~~ bash
curl -O https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/kubernetes_export_sa.sh && sed -i 's/kubectl/microk8s kubectl/g' \
kubernetes_export_sa.sh && bash kubernetes_export_sa.sh cd-user \
devtroncd https://raw.githubusercontent.com/devtron-labs/utilities/main/kubeconfig-exporter/clusterrole.yaml
~~~

{% endtab %}
{% endtabs %}

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/generate-cluster-credentials.png)
### Vorteile einer selbst gehosteten URL
* Notfallwiederherstellung:
  * Es ist nicht möglich, die Server-URL eines bestimmten Cloud-Anbieters zu bearbeiten. Falls Sie eine EKS URL verwenden (z. B.` *****.eu-west-1.elb.amazonaws.com`), erweist sich das Hinzufügen eines neuen Clusters und die nacheinander erfolgende separate Migration aller Dienste als sehr mühselige Aufgabe.
  * Mit einer selbst gehosteten URL dagegen (e.g. `clear.example.com`) können Sie im DNS-Manager einfach auf die Server-URL des neuen Clusters verweisen, das neue Cluster-Token aktualisieren und alle Bereitstellungen synchronisieren.
* Einfache Cluster-Migrationen:
  * Bei verwalteten Kubernetes Clustern (etwa EKS, AKS, GKE etc.), die für einen jeweiligen Cloud-Provider spezifisch sind, verursacht die Migration eines Clusters von einem Anbieter zu einem anderen eine enorme Zeit- und Kraftverschwendung.
  * Dagegen ist andererseits die Migration für eine selbst gehostete URL einfach, da die URL eine einzige gehostete Domäne und unabhängig vom Cloud-Anbieter ist.
### Konfigurieren Sie Prometheus (Aktivieren Sie Anwendungsmetriken)
Falls Sie Anwendungsmetriken für die im Cluster zur Verfügung stehenden Anwendungen anzeigen möchten, muss Prometheus im Cluster bereitgestellt werden. Prometheus bietet als leistungsstarkes Werkzeug einen grafischen Einblick in das Verhalten Ihrer Anwendung.
> **Hinweis:** Stellen Sie sicher, dass Sie `Monitoring (Grafana)` vom `Devtron Stack Manager` installieren, um Prometheus konfigurieren zu können.
> Sollten Sie `Monitoring (Grafana)` nicht installieren, steht die Option einer Konfiguration von Prometheus nicht zur Verfügung.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/enable-app-metrics.png)

Aktivieren Sie die Anwendungsmetriken, um Prometheus zu konfigurieren, und tragen Sie die Informationen in die folgenden Felder ein:

|Feld|Beschreibung|
| :- | :- |
|`Prometheus Endpunkt`|Geben Sie die URL Ihres Prometheus an:|
|`Authentifizierungstyp`|Prometheus unterstützt zwei Authentifizierungstypen:<ul><li>**Basic:** Wenn Sie den Authentifizierungstyp `Basic` wählen, müssen Sie zur Authentifizierung den `Benutzernamen` und das `Passwort` für Prometheus angeben.</li></ul> <ul><li>**Anonymous:** Wenn Sie den Authentifizierungstyp `Anonymous` wählen, müssen Sie weder den `Benutzernamen` noch das `Passwort` angeben.<br>Hinweis: Die Felder `Benutzername` und `Passwort` werden standardmäßig nicht verfügbar sein.</li></ul>|
|`TLS-Schlüssel` & `TLS-Zertifikat`|Der `TLS-Schlüssel` und das `TLS-Zertifikat` sind optional. Diese Optionen werden eingesetzt, wenn Sie eine angepasste URL verwenden.|

Klicken Sie bitte jetzt auf `Save Cluster`, um Ihren Cluster auf Devtron zu speichern.
### Devtron Agent installieren
Ihr Kubernetes Cluster wird mit Devtron kartiert, wenn Sie die Cluster-Konfigurationen speichern. Als Nächstes muss der Devtron-Agent auf dem ergänzten Cluster installiert werden, damit Sie Ihre Anwendungen auf diesem Cluster bereitstellen können.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/install-devtron-agent.png)

Klicken Sie zu Beginn der Devtron-Installation auf `Details`, um den Status der Installation zu kontrollieren.

![Install Devtron Agent](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-agents.jpg)

Das sich nun neu öffnende Fenster zeigt alle Details des Devtron Agenten an.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/cluster\_gc5.jpg)
## Umgebung hinzufügen
Haben Sie Ihren Cluster in `Clusters & Environments` hinterlegt, können Sie eine Umgebung hinzufügen, indem Sie `Add environment` anklicken.

Ein "Neue Umgebung"-Fenster öffnet sich.

|Feld|Beschreibung|
| :- | :- |
|`Name der Umgebung`|Tragen Sie den Namen Ihrer Umgebung ein.|
|`Tragen Sie einen Namespace (Namensraum) ein`|Tragen Sie einen Ihrer Umgebung entsprechenden Namespace ein.<br>**Hinweis**: Falls dieser Namespace in Ihrem Cluster noch nicht existiert, wird Devtron ihn anlegen. Falls er bereits vorhanden ist, wird Devtron die Umgebung dem Namespace zuordnen.</br>|
|`Umgebungstyp`|Wählen Sie Ihren Umgebungstyp:<ul><li>`Production`</li></ul> <ul><li>`Non-Production`</li></ul>Hinweis: Devtron zeigt Verteilungsmetriken (DORA-Metriken) ausschließlich für Umgebungen an, die als `Produktion` gekennzeichnet sind.|

Klicken Sie auf `Save` und Ihre Umgebung wird erstellt.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/cluster-and-environments/gc-cluster-add-environment.jpg)
## Umgebung aktualisieren
* Außerdem können Sie eine Umgebung aktualisieren, indem Sie die Umgebung anklicken.
* Sie können nur die Optionen `Production` und `Non-Production` ändern.
* Nicht veränderbar sind `Environment Name` und `Namespace Name`.
* Versäumen Sie nicht, auf **Update** zu klicken, um Ihre Umgebung zu aktualisieren.
