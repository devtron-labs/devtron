# Erste Schritte
Dieser Abschnitt umfasst Informationen hinsichtlich der zu erfüllenden Mindestanforderungen für die Installation und den Einsatz von **Devtron**.

Devtron wird über einen Kubernetes-Cluster installiert. Sobald Sie ein Kubernetes-Cluster erstellt haben, kann Devtron eigenständig für sich installiert werden oder mit CI/CD-Integration:

* [Devtron mit CI/CD](setup/install/install-devtron-with-cicd.md): Die Devtron-Installation zusammen mit der CI/CD-Integration dient zur Ausführung von CI/CD, Sicherheitsscans, GitOps, Debugging und Beobachtbarkeit.
* [Helm Dashboard von Devtron](setup/install/install-devtron.md): Das Helm-Dashboard von Devtron, eine eigenständige Installation, umfasst Funkionalitäten, um in einer Vielzahl von Clustern vorhandene Helm-Anwendungen einzusetzen, zu beobachten, zu verwalten und zu debuggen. Sie können Integrationen auch vom [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=) aus installieren.

In diesem Abschnitt erläutern wir Ihnen die grundlegenden Aspekte für den schnellen Einstieg in die Arbeit mit **Devtron**.
Lassen Sie uns zuerst auf die Voraussetzungen eingehen, die erfüllt sein müssen, bevor Sie Devtron installieren können.
## Vorgeschriebene Anforderungen
* Erstellen Sie ein [Kubernetes-Cluster, bevorzugt K8s Version 1.16 oder später](#create-a-kubernetes-cluster)
* [Helm-Installation](https://helm.sh/docs/intro/install/)
* [Empfohlene Ressourcen](#recommended-resources)
### Erstellen Sie einen Kubernetes-Cluster
Sie können für die Installation von Devtron irgendein [Kubernetes-Cluster](https://kubernetes.io/docs/tutorials/kubernetes-basics/create-cluster/) erstellen (vorzugsweise K8s Version 1.16 oder später).

Sie können jeweils Ihrem Bedarf entsprechend mit einem der folgenden Anbieter ein Cluster erstellen:

|Cloud-Anbieter|Beschreibung|
| :-: | :-: |
|**AWS EKS**|Ein Cluster mit [AWS EKS](https://docs.aws.amazon.com/eks/latest/userguide/getting-started-console.html) erstellen. <br>`Hinweis`: Nutzen Sie als Referenz zur Installation von `Devtron mit CI/CD` auf AWS EKS gerne unsere angepasste Dokumentation [hier](setup/install/install-devtron-on-AWS-EKS.md). </br>|
|**Google Kubernetes Engine (GKE)**|Ein Cluster mit [GKE](https://cloud.google.com/kubernetes-engine/) erstellen.|
|**Azure Kubernetes Service (AKS)**|Ein Cluster mit [AKS](https://learn.microsoft.com/en-us/azure/aks/) erstellen.|
|**k3s – Lightweight Kubernetes**|Ein Cluster mit [k3s – Lightweight Kubernetes](https://devtron.ai/blog/deploy-your-applications-over-k3s-lightweight-kubernetes-in-no-time/) erstellen. <br>`Hinweis`: Nutzen Sie als Referenz zur Installation von `Helm Dashboard von Devtron` auf `Minikube, Microk8s, K3s, Kind` gerne unsere angepasste Dokumentation [hier](setup/install/Install-devtron-on-Minikube-Microk8s-K3s-Kind.md). </br>|

### Helm installieren
Stellen Sie sicher, dass [Helm](https://helm.sh/docs/intro/install/) installiert wird.
### Empfohlene Ressourcen
Nachfolgend finden Sie die Mindestanforderungen für die Installation von `Helm Dashboard von Devtron` und `Devtron mit CI/CD` entsprechend der Anzahl an Anwendungen, die Sie auf `Devtron` verwalten möchten:

* Für die Konfiguration kleiner Ressourcen (nur zur Verwaltung von bis zu 5 Anwendungen auf Devtron):

|Integration|CPU|Speicher|
| :-: | :-: | :-: |
|**Devtron mit CI/CD**|2|6 GB|
|**Helm Dashboard von Devtron**|1|1 GB|

* Für die Konfiguration mittelgroßer/größerer Ressourcen (zur Verwaltung von mehr als 5 Anwendungen auf Devtron):

|Integration|CPU|Speicher|
| :-: | :-: | :-: |
|**Devtron mit CI/CD**|6|13 GB|
|**Helm Dashboard von Devtron**|2|3 GB|

> Weiterführende Information finden Sie im Abschnitt [Konfigurationen überschreiben](setup/install/override-default-devtron-installation-configs.md).

> **Hinweis:**

* Bitte vergewissern Sie sich, dass die empfohlenen Ressourcen auf Ihrem Kubernetes-Cluster zur Verfügung stehen, bevor Sie mit der Devtron-Installation fortfahren.
* Wir raten davon ab, zur Sicherstellung einer gleichmäßigen Leistung Burstable-CPU-VMs (T-Serie in AWS, B-Serie in Azure und E2/N1 in GCP) für die Devtron-Installation zu verwenden.
## Die Installation von Devtron
Sie können Devtron eigenständig als Standalone installieren (Helm Dashboard von Devtron) oder zusammen mit CI/CD-Integration. Oder Sie können Devtron auch auf die aktuelle Version upgraden.

Wählen Sie entsprechend Ihres Bedarfs eine der Optionen aus:

|Installationsoptionen|Beschreibung|
| :-: | :-: |
|[Devtron mit CI/CD](setup/install/install-devtron-with-cicd.md)|Die Devtron-Installation zusammen mit der CI/CD-Integration wird eingesetzt für CI/CD, Sicherheitsscans, GitOps, Debugging und Beobachtbarkeit.|
|[Helm Dashboard von Devtron](setup/install/install-devtron.md)|Das Helm-Dashboard von Devtron, eine eigenständige Installation, umfasst Funkionalitäten, um in einer Vielzahl von Clustern vorhandene Helm-Anwendungen einzusetzen, zu beobachten, zu verwalten und zu debuggen. Sie können Integrationen auch vom [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=) aus installieren.|
|[Devtron mit CI/CD sowie GitOps (Argo CD)](setup/install/install-devtron-with-cicd-with-gitops.md)|Diese Option gestattet es, Devtron mit CI/CD zu installieren, indem Sie GitOps während der Installation aktivieren. Sie können vom [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations?q=) aus auch andere Integrationen installieren.|
|**Devtron auf die aktuelle Version upgraden**|Ein Upgrade für Devtron kann mittels einer der folgenden Methoden durchgeführt werden: <ul><li>[Upgrade von Devtron über Helm](https://docs.devtron.ai/v/v0.5/getting-started/upgrade#upgrade-devtron-using-helm)</ul></li><ul><li>[Upgrade von Devtron über das UI ](https://docs.devtron.ai/v/v0.5/getting-started/upgrade/upgrade-devtron-ui)</ul></li>|

**Hinweis**: Bitte informieren Sie uns über unseren Discord-Kanal, falls Sie Fragen haben. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
