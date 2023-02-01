# Git-Konten
Git-Konten erlauben es Ihnen, Ihre Code-Quelle mit Devtron zu verbinden. Sie können unter Verwendung dieser Git-Konten mithilfe der CI-Pipeline den Code erstellen
## Git-Konto hinzufügen
Zum Hinzufügen des Git-Kontos gehen Sie bitte zu `Git accounts` im Abschnitt `Global Configurations`. Klicken Sie auf **Add git account** (Git-Konto hinzufügen).

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/git-accounts.jpg)

Tragen Sie zum Hinzufügen Ihres Git-Kontos in den folgenden Feldern die entsprechenden Informationen ein:

|Feld|Beschreibung|
| :- | :- |
|`मुं`|Geben Sie Ihrem Git-Provider einen Namen.<br>Hinweis: Dieser Name wird im > [Git-Repository](../creating-application/git-material.md) auf der Dropdown-Liste der App-Konfiguration zur Verfügung stehen.</br>|
|`Git-Host`|Dies ist der Git-Provider, auf dem das jeweilige Git-Repository der Anwendung gehostet wird.<br>Hinweis: `Bitbucket` und `GitHub` stehen standardmäßig in der Dropdown-Liste zur Verfügung. Sie können so viele Worte ergänzen, wie Sie wollen, indem Sie `[+ Add Git Host]` anklicken.</br>|
|`URL`|Geben Sie die `URL`des Git-Hosts an.<br>Zum Beispiel: <https://github.com> für Github, <https://gitlab.com> für GitLab etc.|
|`Authentifizierungstyp`|Devtron unterstützt drei Authentifizierungstypen:<ul><li>**User auth:** Bei Auswahl von `User auth` als Authentifizierungstyp müssen Sie den `Usernamen` und ein `Password`oder einen `Auth token` zur Authentifizierung Ihres Versionskontrollkontos bereitstellen.</li></ul> <ul><li>**Anonym:** Bei Auswahl von `Anonymous` als Authentifizierungstyp benötigen Sie weder einen `Username` noch ein `Password`.<br>Hinweis: Falls als Authentifizierungstyp `Anonymous` gewählt wurde, besteht ausschließlich Zugang zum öffentlichen Git-Repository.</li></ul><ul><li>**SSH Key:** Bei Auswahl von `SSH Key` als Authentifizierungstyp müssen Sie den `Private SSH Key` bereitstellen, der dem Public Key in Ihrem Versionskontrollkonto hinzugefügt wurde.</li></ul>|

## Git-Konto aktualisieren
Zum Aktualisieren des Git-Kontos:

1. Klicken Sie auf das Git-Konto, das Sie aktualisieren wollen.
1. Aktualisieren Sie die erforderlichen Änderungen.
1. Klicken Sie auf `Update` (Aktualisieren), um die Änderungen zu speichern.

Aktualisierungen können ausschließlich innerhalb eines Authentifizierungstyps oder eines Protokolltyps durchgeführt werden, z. B. HTTPS (Anonym oder Benutzerauthentifizierung) & SSH. Sie können Updates durchführen von `Anonymous` auf `User Auth` und anders herum, aber nicht von `Anonym` oder `User Auth` auf `SSH` bzw. anders herum.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/global-configurations/git-accounts/update-git-accounts.jpg)

Hinweis:

* Sie können ein Git-Konto sowohl aktivieren als auch deaktivieren. Aktivierte Git-Konten sind verfügbar unter App-Konfiguration > [Git repository](../creating-application/git-material.md).

![](../../user-guide/global-configurations/images/git-account-enable-disable.jpg)
