
\# FAQ

<details>

<summary>1. Woher weiß ich, wann die Installation abgeschlossen ist?</summary>

Führen Sie den folgenden Befehl aus, um den Status der Installation zu überprüfen:

\```bash

kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'

\```

Der obige Befehl gibt `Applied` aus, sobald der Installationsvorgang abgeschlossen ist. Der Installationsvorgang kann bis zu 30 Minuten dauern.

</details>

<details>

<summary>2. Wie kann ich den Fortschritt der Installation verfolgen?</summary>

Führen Sie den folgenden Befehl aus, um die Logs des Pods zu überprüfen:

\```bash

pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod

\```

</details>

<details>

<summary>3. Wie kann ich die Installation neu starten, wenn die Logs des Devtron-Installers einen Fehler enthalten?</summary>

Führen Sie zunächst den folgenden Befehl aus, um die vom Devtron-Installer installierten Komponenten zu bereinigen:

\```bash

cd devtron-installation-script/

kubectl delete -n devtroncd -f yamls/

kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'

\```

Danach [install Devtron](./install-devtron.md)

</details>


Wenn Sie immer noch Probleme haben, wenden Sie sich bitte an uns auf [Discord] (https://discord.gg/jsRG5qx2gp).
