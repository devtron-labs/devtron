### Install Devtron using kubectl

If you don't want to install helm and just want to use `kubectl` to install `devtron platform`, then please follow the steps mentioned below:

```bash
git clone https://github.com/devtron-labs/devtron-installation-script.git
cd devtron-installation-script/
#Use a preferred editor to edit the values in install/devtron-operator-configs.yaml
vim install/devtron-operator-configs.yaml
```
Edit the `install/devtron-operator-configs.yaml` to configure your Devtron installation. For more details about it, see [configuration](#configuration)
Once your configurations are ready, continue with following steps
```bash
kubectl create ns devtroncd
kubectl -n devtroncd apply -f charts/devtron/crds
# wait for crd to install
kubectl apply -n devtroncd -f charts/devtron/templates/install.yaml
#edit install/devtron-operator-configs.yaml and input the
kubectl apply -n devtroncd -f install/devtron-operator-configs.yaml
kubectl apply -n devtroncd -f charts/devtron/templates/devtron-installer.yaml
```

