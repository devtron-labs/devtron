\# Devtron installieren


Devtron wird über ein Kubernetes-Cluster installiert. Sobald Sie einen Kubernetes-Cluster erstellt haben, kann Devtron eigenständig oder zusammen mit der CI/CD-Integration installiert werden.

Wählen Sie eine der Optionen je nach Ihren Anforderungen:

| Installationsoptionen | Beschreibung | Wann auswählen |

\| --- | --- | --- |

| [Devtron with CI/CD](https://docs.devtron.ai/install/install-devtron-with-cicd) | Die Devtron-Installation zusammen mit der CI/CD-Integration dient zur Ausführung von CI/CD, Sicherheitsscans, GitOps, Debugging und Beobachtbarkeit. | Verwenden Sie diese Option, um Devtron mit der `Build and Deploy CI/CD`-Integration zu installieren. |

| [Helm Dashboard by Devtron](https://docs.devtron.ai/install/install-devtron) | Das Helm Dashboard von Devtron ist eine eigenständige Installation mit Funktionen für den Einsatz, die Beobachtung, die Verwaltung und das Debugging bestehender Helm-Anwendungen in mehreren Clustern. Sie können auch Integrationen von [Devtron Stack Manager] (https://docs.devtron.ai/v/v0.6/usage/integrations) installieren. | Verwenden Sie diese Option, wenn Sie die Anwendungen über Helm verwalten und Devtron zur Bereitstellung, Beobachtung, Verwaltung und Fehlersuche der Helm-Anwendungen verwenden möchten. |

| [Devtron with CI/CD along with GitOps (Argo CD)](https://docs.devtron.ai/install/install-devtron-with-cicd-with-gitops) | Mit dieser Option können Sie Devtron mit CI/CD installieren, indem Sie GitOps während der Installation aktivieren. Sie können auch andere Integrationen von [Devtron Stack Manager](https://docs.devtron.ai/v/v0.6/usage/integrations) installieren. | Verwenden Sie diese Option, um Devtron mit CI/CD zu installieren, indem Sie GitOps aktivieren. Das ist die skalierbarste Methode in Bezug auf Versionskontrolle, Zusammenarbeit, Compliance und Infrastrukturautomatisierung.  |


\*\*Hinweis\*\*: Wenn Sie Fragen haben, melden Sie sich bitte in unserem Discord-Channel. [![Join Discord](https://img.shields.io/badge/Join%20us%20on-Discord-e01563.svg)](https://discord.gg/jsRG5qx2gp)
