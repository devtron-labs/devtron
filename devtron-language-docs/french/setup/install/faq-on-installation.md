
\# FAQ

<details>

<summary>1. Comment saurais-je que l'installation est terminée ?</summary>

Exécutez la commande suivante pour vérifier le statut de l'installation :

\```bash

kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'

\```

La commande ci-dessus indiquera `Applied` une fois le processus d'installation terminé. Le processus d'installation peut prendre jusqu'à 30 minutes.

</details>

<details>

<summary>2. Comment puis-je suivre la progression de l'installation ? </summary>

Exécutez la commande suivante pour vérifier les journaux du Pod :

\```bash

pod=$(kubectl -n devtroncd get po -l app=inception -o jsonpath='{.items[0].metadata.name}')&& kubectl -n devtroncd logs -f $pod

\```

</details>

<details>

<summary>3. Comment puis-je redémarrer l'installation si les journaux du programme d'installation Devtron contiennent une erreur ?</summary>

Exécutez d'abord la commande ci-dessous pour nettoyer les composants installés par le programme d'installation Devtron :

\```bash

cd devtron-installation-script/

kubectl delete -n devtroncd -f yamls/

kubectl -n devtroncd patch installer installer-devtron --type json -p '[{"op": "remove", "path": "/status"}]'

\```

Ensuite, [install Devtron](./install-devtron.md)

</details>


Si vous rencontrez toujours des problèmes, veuillez nous contacter sur [Discord](https://discord.gg/jsRG5qx2gp).
