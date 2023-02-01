# GitOps
Devtron verwendet GitOps, um die Kubernetes-Konfigurationsdateien der Anwendungen zu speichern.
Zum Speichern der Konfigurationsdateien und des gewünschten Zustands der Anwendungen sind die Git Anmeldedaten unter **Global Configurations** > **GitOps** im Devtron Dashboard einzutragen.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/git-providers.jpg)

Im Folgenden finden Sie die in Devtron verfügbaren Git-Provider. Wählen Sie einen der Git-Providerer (z. B. GitHub), um GitOps zu konfigurieren:

* [GitHub](#github)
* [GitLab](#gitlab)
* [Azure](#azure)
* [BitBucket Cloud](#bitbucket-cloud)

**Hinweis**: Der von Ihnen zur GitOps-Konfiguration gewählte Git-Provider wirkt auf folgende Abschnitte ein:

* Deployment Template, [klicken Sie hier](https://docs.devtron.ai/user-guide/creating-application/deployment-template) für weitere Informationen.
* Diagramme, [klicken Sie hier](https://docs.devtron.ai/user-guide/deploy-chart) für weitere Informationen.
## GitHub
Falls sie `GitHuB` als Ihren Git-Provider auswählen, tragen Sie zum Konfigurieren von GitOps von bitte die Informationen in die folgenden Felder ein:

|Felder|Beschreibung|
| :-: | :-: |
|**Git-Host**|In diesem Feld wird die URL des ausgewählten Git-Anbieters angezeigt. <br>Zum Beispiel: https://github.com/ for GitHub.</br>|
|**GitHub Organisation Name**|Tragen Sie den Namen der GitHub-Organisation ein.<br>Falls Sie noch keinen haben, erstellen Sie diesen bitte mittels[ how to create organization in Github](#how-to-create-organization-in-github) (Anleitung zum Erstellen einer Organisation in GitHub).</br>|
|**GitHub Benutzername**|Tragen sie den Benutzernamen Ihres GitHub-Kontos ein.|
|**Personal Access Token (Persönlicher Zugangstoken)**|Geben Sie ihren Personal Access Token (PAT) an. Er findet Verwendung als Alternative zum Passwort, um Ihr GitHub-Konto zu authentifizieren.<br>Falls Sie keinen haben, können Sie ein GitHub PAT [hier](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) erstellen.</br>|

### Wie man in GitHub eine Organisation erstellt
**Hinweis**: Es ist nicht ratsam, die GitHub-Organisation zu verwenden, die Ihren Quellcode enthält.

{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/github/github-gitops-latest.mp4" caption="GitHub" %}

1. Erstellen Sie ein neues Konto auf GitHub (für den Fall, dass Sie noch keines haben).
1. Klicken Sie in der oberen rechten Ecke Ihrer GitHub-Seite auf Ihr Profilfoto und dann auf `Settings`.
1. Klicken Sie im Bereich `Access` auf `Organizations`.
1. Klicken Sie im Bereich `Organizations` auf `New organization`.
1. Wählen Sie einen [Plan](https://docs.github.com/en/get-started/learning-about-github/githubs-products) für Ihre Organisation aus. Sie können auch `create free organization` auswählen.
1. Tragen Sie auf der Seite `Set up your organization`
   * bitte `den Kontonamen der Organisation ein und` `die Kontakt-E-Mail`.
   * Wählen Sie die Option, zu der Ihre Organisation gehört.
   * Überprüfen Sie Ihr Konto und klicken Sie auf `Next`.
   * `Der Name `Ihrer `GitHub-Organisatione` wird erstellt.
1. Gehen Sie zu Ihrem Profil und klicken Sie auf `Your organizations` für eine Ansicht aller von Ihnen erstellten Organisationen.

Weitere Informationen über die Ihrem Team zur Verfügung stehenden Pläne finden sie auf [GitHub's products](https://docs.github.com/en/get-started/learning-about-github/githubs-products). Daneben können Sie auch die offizielle Seite der [GitHub Organisation](https://docs.github.com/en/github/setting-up-and-managing-organizations-and-teams/about-organizations) für weiterführende Details einsehen.

**Hinweis**:

* repo – Vollständige Kontrolle über private Repositories (möglicher Zugriff auf Commit-Status, Deployment-Status und öffentliche Repositories).
* admin:org – Volle Kontrolle über Organisationen und Teams (Lese- und Schreibzugriff).
* delete\_repo – Gestattet Zugriff auf private Repositories, um sie zu löschen.
## GitLab
Falls Sie `GitLab` als Ihren Git-Provider auswählen, tragen Sie bitte die Informationen in den folgenden Feldern ein, um GitOps zu konfigurieren:

|Felder|Beschreibung|
| :-: | :-: |
|**Git-Host**|Dieses Feld zeigt die URL des gewählten Git-Providers an. <br>Zum Beispiel:<br>https://gitlab.com/ for GitLab.|
|**GitLab Gruppen-ID**|Tragen Sie die GitLab-Gruppen-ID ein.<br>Falls Sie noch keine haben, erstellen Sie bitte eine mittels [GitLab Group ID](#how-to-create-organization-in-gitlab).</br>|
|**GitLab Benutzername**|Tragen Sie den Benutzernamen Ihres GitLab-Kontos ein.|
|**Personal Access Token (Persönlicher Zugangstoken)**|Geben Sie ihren Personal Access Token (PAT) an. Es findet als Alternative zum Passwort Verwendung, um Ihr GitLab-Konto zu authentifizieren.<br>Falls Sie noch keines haben, können Sie [hier](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html) ein GitLab-PAT erstellen.</br>|

### Wie man in GitLab eine Organisation erstellt
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/gitlab/gitops-gitlab-latest1.mp4" caption="GitHub" %}

1. Erstellen Sie ein neues Konto bei GitLab (sofern Sie noch keines haben).
1. Sie können eine Gruppe erstellen, in dem Sie auf dem GitLab-Dashboard zur Registerkarte „Gruppen‟ gehen und `New group` anklicken.
1. Wählen Sie `Create group`.
1. Tragen Sie den Gruppenamen ein (Pflichtfeld), wählen Sie optional die Beschreibungen aus und klicken Sie auf `Create group` (Gruppe erstellen).
1. Ihre Gruppe wird erstellt und Ihr Gruppenname erhält eine neue `Gruppen-ID` (z. B. 61512475).

**Hinweis**:

* api – Gewährt vollständigen Lese-/Schreibzugriff auf die Projekt-API mit Gültigkeitsbereich
* write\_repository – Erlaubt Lese- und Schreibzugriff (pull, push) auf das Repository
## Azure
Wenn Sie `GitAzureLab` als Ihren Git-Provider auswählen, tragen Sie bitte die Informationen in den folgenden Feldern ein, um GitOps zu konfigurieren:

|Felder|Beschreibung|
| :-: | :-: |
|**Azure DevOps Organisation Url**\*|Dieses Feld zeigt die URL des gewählten Git-Providers an. <br>Zum Beispiel:<br>https://dev.azure.com/ for Azure.|
|**Azure DevOps Projektname**|Tragen Sie den Azure DevOps Projektnamen ein.<br>Falls Sie keinen haben, können Sie einen mit [Azure DevOps Project Name](#how-to-create-azure-devops-project-name) erstellen.</br>|
|**Azure DevOps Benutzername**\*|Tragen Sie de Benutzernamen Ihres Azure DevOps-Kontos ein.|
|**Azure DevOps Zugangstoken**\*|Geben Sie Ihren Azure DevOps Zugangstoken an. Er findet als Alternative zum Passwort Verwendung, um Ihr Azure DevOps-Konto zu authentifizieren.<br>Falls Sie keinen haben, können Sie [hier](https://docs.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=preview-page)einen Azure DevOps Zugangstoken erstellen.</br>|

### Wie man einen Azure DevOps Projektnamen erstellt
**Hinweis**: Sie benötigen eine Organisation, bevor Sie ein Projekt erstellen können. Falls Sie noch keine Organisation erstellt haben, befolgen Sie bitte die Anweisungen unter [Anmelden, bei Azure DevOps einloggen](https://learn.microsoft.com/en-us/azure/devops/user-guide/sign-up-invite-teammates?view=azure-devops), wodurch ebenfalls ein Projekt erstellt wird. Oder siehe [Erstellen einer Organisation oder Projektsammlung](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/create-organization?view=azure-devops).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-new-project.jpg)

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/azure+devops/azure-create-new-project.jpg)

1. Gehen Sie zu Azure DevOps und navigieren Sie zu den Projekten.
1. Wählen Sie Ihre Organisation und klicken Sie auf `New project` (Neues Projekt).
1. Auf der Seite `Create new project` (Neues Projekt erstellen),
   * tragen Sie den `Projektnamen` und die Beschreibung des Projekts ein.
   * Wählen Sie die Sichtbarkeitsoption aus (privat oder öffentlich), den anfänglichen Versionskontrolltyp und den Work-Item-Prozess.
   * Klicken Sie dann auf `Create`.
   * Azure DevOps zeigt nun die Projekt-Startseite mit dem `Projektnamen` an.

Sie können für weitere Informationen auch auf die offizielle Dokumentenseite [Azure DevOps Project Name](https://docs.microsoft.com/en-us/azure/devops/organizations/projects/create-project?view=azure-devops&tabs=preview-page) wechseln.

**Hinweis**:

* code – Gestattet das Lesen des Quellcodes und der Metadaten zu Commits, Change Sets, Branches und anderen Artefakten der Versionskontrolle. [Mehr Informationen zu Bereichen in Azure devops](https://docs.microsoft.com/en-us/azure/devops/integrate/get-started/authentication/oauth?view=azure-devops#scopes).
## BitBucket Cloud
Wenn Sie `Bitbucket Cloud` als Ihren Git-Provider wählen, tragen Sie bitte die Informationen in den folgenden Feldern ein, um GitOps zu konfigurieren:

|Felder|Beschreibung|
| :-: | :-: |
|**Bitbucket Host**|Dieses Feld zeigt die URL des gewählten Git-Providers an. <br>Zum Beispiel:<br>https://bitbucket.org/ for Bitbucket.|
|**Bitbucket Workspace ID**|Tragen Sie die Bitbucker workspace ID ein.<br>Falls Sie keine haben, können sie mittels [Bitbucket Workspace Id](#how-to-create-bitbucket-workspace-id) eine erstellen.</br>|
|**Bitbucket Project Key (Projektschlüssel)**|Tragen Sie den Bitbucket Projektschlüssel.<br>Falls Sie keinen haben, können Sie mittels [Bitbucket Project Key](https://support.atlassian.com/bitbucket-cloud/docs/group-repositories-into-projects/) einen erstellen.</br><br>Hinweis: Dieses Feld ist kein Pflichtfeld. Sollte das Projekt nicht angezeigt werden, wird das Repository automatisch dem ältesten Projekt im Arbeitsbereich zugewiesen.</br>|
|**Bitbucket Benutzername**\*|Tragen Sie den Benutzernamen Ihres Bitbucket-Kontos ein.|
|**Personal Access Token (Persönlicher Zugangstoken)**|Geben Sie ihren Personal Access Token (PAT) an. Er findet Verwendung als Alternative zum Passwort, um Ihr Bitbucket Cloud-Konto zu authentifizieren.<br>Falls Sie keines haben, können Sie [hier](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/) ein Bitbucket Cloud PAT erstellen.</br>|

### Wie man eine Bitbucket Workspace ID erstellt
{% embed url="https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/gitops/bitbucket/bitbucket-latest-gitops.mp4" caption="GitHub" %}

1. Richten Sie ein neues individuelles Konto bei Bitbucket ein (sofern Sie noch keines haben).
1. Wählen Sie in der oberen rechten Ecke der oberen Navigationsleiste Ihren Profil- und Einstellungsavatar aus.
1. Wählen Sie im Dropdown-Menu `All workspaces`.
1. Wählen Sie `Create workspace` in der oberen rechten Ecke der Seite `Workspaces`.
1. Auf der Seite `Create a Workspace`:
* Tragen Sie einen `Workspace Namen` ein.
* Tragen Sie eine `Workspace ID` ein. Ihre ID darf keine Leer- oder Sonderzeichen enthalten. Zahlen und Großbuchstaben sind erlaubt. Diese ID wird Teil der URL für den Workspace (Arbeitsbereich) und für jeden Bereich, wo es eine Kennzeichnung gibt, die das Team identifiziert (APIs, Berechtigungsgruppen, OAuth, usw.).
* Klicken Sie dann auf `Create`.
6. Ihr `Workspace-Name` und die `Workspace-ID` werden erstellt.

Sie können für weiterführende Details auch zur offiziellen Dokumentenseite [Bitbucket Workspace Id](https://support.atlassian.com/bitbucket-cloud/docs/what-is-a-workspace/) wechseln.

**Hinweis**:

* repo – Volle Kontrolle über Repositories (Lesen, Schreiben, Verwalten, Löschen).

Klicken Sie auf **Save**, um Ihre GitOps-Konfigurationsdetails zu speichern.

**Hinweis**: Der aktivierte GitOp-Provider ist mit einem grünen Häkchen markiert.
