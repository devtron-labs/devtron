# Override Default Configurations of Devtron Installation

Installing Devtron is quite a straightforward process with the default installation configurations. Devtron supports installation using [helm3](install-devtron-helm-3.md), [helm2](install-devtron-helm-2.md), and [kubectl](install-devtron-using-kubectl.md). To tweak the default configurations and make changes as per our requirements, we need to create our own yaml files with some standard values and specs to make necessary changes. 


## Installation

### Installing from Scratch
When we are installing Devtron from scratch and want to give our own custom yaml files to override the default configurations, we have to first create a namespace and then we can apply changes to the files which we want to override. Please execute the following command to create a namesapce in your cluster where Devtron will be installed - 

```bash
kubectl create namespace devtroncd
```

As the namepsace is being created, we need to understand how to make changes in the existing yaml files or create a new one to override the default configurations. The existing templates which needs to be modified for overriding configs are available at [updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates). 

Let's take an example, create a new `devtron-argocd-override.yaml` to override default argocd configuration while installing Devtron and understand the template structure for editing the file or creating a new yaml file from scratch. 

In the [installation-script](https://github.com/devtron-labs/devtron/blob/main/manifests/installation-script) from line 127 to 151 we can see the names which will be used in the metadata of our corresponding yaml file for overriding configs. 

```bash
######Downloading the manifests

argocdOverride = kubectl get cm -n devtroncd argocd-override-cm;
clairOverride = kubectl get cm -n devtroncd clair-override-cm;
clairConfigOverride = kubectl get cm -n devtroncd clair-config-override-cm;
dashboardOverride = kubectl get cm -n devtroncd dashboard-override-cm;
gitSensorOverride = kubectl get cm -n devtroncd git-sensor-override-cm;
guardOverride = kubectl get cm -n devtroncd guard-override-cm;
postgresqlOverride = kubectl get cm -n devtroncd postgresql-override-cm;
imageScannerOverride = kubectl get cm -n devtroncd image-scanner-override-cm;
kubewatchOverride = kubectl get cm -n devtroncd kubewatch-override-cm;
lensOverride = kubectl get cm -n devtroncd lens-override-cm;
```

To apply new configs for our argocd, create a new file named `devtron-argocd-override.yaml` in [updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates) directory as discussed above. From the above code we can see the name `argocd-override-cm` which would be used in the metadata of corresponding yaml file which is `devtron-argocd-override.yaml`. In the yaml file, there are some standards which need to be followed while creating any new yaml file for overriding default configs.

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  name: corresponding_name_from_installation-script
  namespace:  devtroncd
data:
  override: |

```

The above lines would be used for every new yaml file. In metadata name, as we are overriding argocd configs, we need to replace it with - `argocd-override-cm`. Now after creating a new file in [updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates), we need to look back into default yaml files which are used while installation. The default yaml files are present in [`yamls`](https://github.com/devtron-labs/devtron/tree/main/manifests/yamls). For argocd, we can observe that we have a file named `argocd.yaml` which contains the default configs which is triggered while installing Devtron. When you open the file, you could see the complete configuration. Let's say you want to override the `Deployment` for `argocd-server` with your own configs. To make this happen, open the your `devtron-argocd-override.yaml` file and update the `argocd-server` - 

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  name: corresponding_name_from_installation-script
  namespace:  devtroncd
data:
  override: |
  	apiVersion: apps/v1
	kind: Deployment
	metadata:
		name: argocd-server
	update:

```

In the above yaml file, you can observe we have just added the `apiVersion`, `kind` and `metadata` name for which we need to update the configurations. After adding the `update:` we are ready to override new configs for `argocd-server`. Let's say we want to add our own custom resources for `argocd-server` - 

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  name: corresponding_name_from_installation-script
  namespace:  devtroncd
data:
  override: |
  	apiVersion: apps/v1
	kind: Deployment
	metadata:
	  name: argocd-server
	update:
	  spec:
    	template:
          spec:
            containers:
              resources:
                limits:
                  cpu: 10m
                  memory: 60Mi
                requests:
                  cpu: 3m
                  memory: 20Mi
  
```

Bingo! We have successfully added our custom resources for `argocd-server` to override argocd default config. But as you can observe, there are many other `Deployment`, `ConfigMap`, `Secret`, `StatefulSet`, etc for other services. To update other services as well, we can simply add the new changes in the same if it `argocd-config`. Let's assume I want to update `Deployment` for `argocd-redis` and update `StatefulSet` of `argocd-application-controller` - 

```bash
apiVersion: v1
kind: ConfigMap
metadata:
  name: corresponding_name_from_installation-script
  namespace:  devtroncd
data:
  override: |
  	apiVersion: apps/v1
	kind: Deployment
	metadata:
	  name: argocd-server
	update:
	  spec:
    	template:
          spec:
            containers:
              resources:
                limits:
                  cpu: 10m
                  memory: 60Mi
                requests:
                  cpu: 3m
                  memory: 20Mi
    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: argocd-redis
    update:
      spec:
        template:
          spec:
            containers:
              resources:
                limits:
                  cpu: 5m
                  memory: 10Mi
                requests:
                  cpu: 1m
                  memory: 2Mi
    ---
    apiVersion: apps/v1
    kind: StatefulSet
    metadata:
      name: argocd-application-controller
    update:
      spec:
        template:
          spec:
            containers:
              resources:
                limits:
                  cpu: 300m
                  memory: 220Mi
                requests:
                  cpu: 100m
                  memory: 100Mi

```

So, we have updated our `devtron-argocd-override.yaml` with `Deployment` for `argocd-redis` and `StatefulSet` of `argocd-application-controller`. In the similar manner we can update any configuration which we want to and for any service be it `argocd` or `clair`. The yaml file standards would remain the same for every service to be overridden and the new files must be placed in [updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates).

Now after creating namespace `devtroncd`, your custom `devtron-argocd-override.yaml` you are ready to apply these changes. For your convinience, we have already provided the templates for all the services which you can go ahead and update as you want in [updates](https://github.com/devtron-labs/devtron/tree/main/manifests/updates). If you have changed all the files present in the directory and want to apply all the yaml files, please execute the following command -

```bash
kubectl apply -f dir -n devtroncd
```
If you want to apply a specific file which you changed, please execute the following command -

```bash
kubectl apply -f file_name -n devtroncd
```

After applying the new configs, you are ready to run the installation script. Please go ahead and continue your installation from your preferred installer - [helm3](install-devtron-helm-3.md), [helm2](install-devtron-helm-2.md) or [kubectl](install-devtron-using-kubectl.md).


### Override new configs if already Installed





