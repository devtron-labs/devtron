# Devtron auf Minikube, Microk8s, K3s, Kind installieren
Sie können Devtron auf einem High-End-Rechner oder auf einer Cloud-VM installieren und ausprobieren. Wenn Sie es auf einem Laptop/PC installieren, kann es anfangen, langsam zu reagieren, daher ist es empfehlenswert, Devtron von Ihrem System zu deinstallieren, bevor Sie es herunterfahren.
## Systemkonfigurationen für die Devtron-Installation
1. 2 vCPUs
1. 4 GB+ freier Arbeitsspeicher
1. 20GB+ freier Festplattenspeicher
## Bevor Sie anfangen
Bevor wir mit der Installation von Devtron beginnen, müssen Sie das Cluster auf Ihrem Server einrichten und die erforderlichen Voraussetzungen installieren:

* Erstellen Sie ein Cluster mit [Minikube](https://minikube.sigs.k8s.io/docs/start/) oder [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) oder [K3s](https://rancher.com/docs/k3s/latest/en/installation/).
* Installieren Sie [Helm3](https://helm.sh/docs/intro/install/).
* Installieren Sie [kubectl](https://kubernetes.io/docs/tasks/tools/).
## Devtron installieren
{% tabs %}

{% tab title=" Minikube/Kind cluster" %}

Um Devtron auf einem ``Minikube/Kind``-Cluster zu installieren, führen Sie den folgenden Befehl aus:
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 
~~~

{% endtab %}

{% tab title="k3s Cluster" %}
Um Devtron auf dem ``k3s``-Cluster zu installieren, führen Sie den folgenden Befehl aus:
~~~ bash
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml

helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort
~~~

{% endtab %}

{% endtabs %}
## Devtron-Dashboard
Führen Sie den folgenden Befehl aus, um auf das Devtron Dashboard zuzugreifen, wenn Sie ``Minikube`` als Cluster verwenden:
~~~ bash
minikube service devtron-service --namespace devtroncd
~~~

Um auf das Devtron-Dashboard zuzugreifen, wenn Sie ``Kind/k3s`` als Cluster verwenden, führen Sie den folgenden Befehl aus, um den Devtron-Dienst auf Port 8000 weiterzuleiten:
~~~ bash
kubectl -ndevtroncd port-forward service/devtron-service 8000:80
~~~

**Dashboard**: <http://127.0.0.1:8000>.
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
## Devtron auf einer Cloud-VM installieren (AWS ec2, Azure VM, GCP VM)
Es wird empfohlen, eine Cloud VM mit 2vCPU+, 4GB+ freiem Arbeitsspeicher, 20GB+ Speicherplatz, rechenoptimiertem VM-Typ & Ubuntu Flavoured OS zu verwenden.
### Microk8s Cluster erstellen
~~~ bash
sudo snap install microk8s --classic --channel=1.22
sudo usermod -a -G microk8s $USER
sudo chown -f -R $USER ~/.kube
newgrp microk8s
microk8s enable dns storage helm3
echo "alias kubectl='microk8s kubectl '" >> .bashrc
echo "alias helm='microk8s helm3 '" >> .bashrc
source .bashrc
~~~
### Devtron installieren
~~~ bash
helm repo add devtron https://helm.devtron.ai

helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd \
--set components.devtron.service.type=NodePort 
~~~
### Führen Sie den folgenden Befehl aus, um die Portnummer des Devtron-Service zu ermitteln:
~~~ bash
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.spec.ports[0].nodePort}'
~~~

Stellen Sie sicher, dass der Port, auf dem der devtron-Dienst läuft, in der Sicherheitsgruppe der VM oder der Netzwerksicherheitsgruppe offen bleibt.

**Hinweis**: Wenn Sie Devtron deinstallieren oder den Devtron Helm-Installer bereinigen möchten, lesen Sie bitte unsere Anleitung zum [Deinstallieren von Devtron](https://docs.devtron.ai/install/uninstall-devtron).

Wenn Sie Fragen haben, melden Sie sich bitte in unserem Discord-Channel. ![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)[](https://discord.gg/jsRG5qx2gp)
