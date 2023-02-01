# Devtron installieren
In diesem Abschnitt beschreiben wir, wie Sie Helm Dashboard von Devtron ohne jegliche Integrationen installieren können. Integrationen können später mit dem [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations) hinzugefügt werden.

Wenn Sie Devtron auf Minikube, Microk8s, K3s, Kind installieren wollen, lesen Sie diesen [Abschnitt](./Install-devtron-on-Minikube-Microk8s-K3s-Kind.md).
## Bevor Sie anfangen
Installieren Sie [Helm](https://helm.sh/docs/intro/install/) falls Sie es noch nicht installiert haben.
## Fügen Sie Help Repo hinzu
~~~ bash
helm repo add devtron https://helm.devtron.ai
~~~
## Installieren Sie Helm Dashboard von Devtron
**Hinweis**: Mit diesem Installationsbefehl wird keine CI/CD-Integration installiert. Für CI/CD lesen Sie den Abschnitt [Devtron mit CI/CD installieren](https://docs.devtron.ai/install/install-devtron-with-cicd).

Führen Sie den folgenden Befehl aus, um Helm Dashboard von Devtron zu installieren:
~~~ bash
helm install devtron devtron/devtron-operator \
--create-namespace --namespace devtroncd
~~~
## Installieren Sie Multi-Architektur-Knoten (ARM und AMD)
Um Devtron auf Clustern mit Multi-Architektur-Knoten (ARM und AMD) zu installieren, fügen Sie dem Devtron-Installationsbefehl diese Option hinzu: `--set installer.arch=multi-arch`.

[//]: # (Wenn Sie planen, Hyperion für `Production Deployments` zu verwenden, beachten Sie bitte unsere empfohlenen Overrides für [Devtron Installation]&#40;override-default-devtron-installation-configs.md&#41;).

[//]: # (## Status der Installation)

[//]: # ()
[//]: # (Den folgenden Befehl ausführen)

[//]: # ()
[//]: # (```bash)

[//]: # (kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}')

[//]: # (```)
## Devtron-Dashboard
Führen Sie den folgenden Befehl aus, um die URL des Dashboards abzurufen:
~~~ text
kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
~~~

Sie erhalten ein Ergebnis wie unten abgebildet:
~~~ text
[test2@server ~]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}'
[map[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com]]
~~~

Der oben genannte Hostname `aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com` ist die Loadbalancer-URL, unter der Sie auf das Devtron-Dashboard zugreifen können.
> Sie können auch einen CNAME-Eintrag vornehmen, der Ihrer Domain/Subdomain entspricht und auf diese Loadbalancer-URL verweist, um über eine benutzerdefinierte Domain darauf zuzugreifen.

|> Host|> Typ|> Verweist auf|
| :- | :- | :- |
|> devtron.yourdomain.com|> CNAME|> aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com|
>
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

**Hinweis**: Wenn Sie Devtron deinstallieren oder den Devtron Helm-Installer bereinigen möchten, lesen Sie bitte unsere Anleitung zum [Deinstallieren von Devtron](https://docs.devtron.ai/install/uninstall-devtron).
## Upgrade
Um die CI/CD-Funktionen mit Devtron zu nutzen, können Sie [Devtron mit CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) oder [Devtron mit CI/CD zusammen mit GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops) installieren.
