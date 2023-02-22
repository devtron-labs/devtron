# Install Devtron Kubernetes Client

Devtron Kubernetes client is an intuitive Kubernetes Dashboard or a command line utility installed outside a Kubernetes cluster. The client can be installed on a desktop running on any Operating Systems and interact with all your Kubernetes clusters and workloads through an API server.
It is a binary, packaged in a bash script that you can download and install by using the following set of commands:


* Download the bash script using the below URL:
https://cdn.devtron.ai/k8s-client/devtron-install.bash

* To automatically download the executable and to open the dashboard in the respective browser, run the following command:

`Note`: Make sure `devtron-install.bash` is placed in the current directory before you run the command.

```bash
sh devtron-install.bash start  
```

* Devtron Kubernetes Client opens in your browser automatically.

## Some Peripheral Commands

* To access the UI of the Kubernetes client, execute the following command. It will open the dashboard through a port in the available web browser and store the Kubernetes client's state.

```bash
sh devtron-install.bash open 
```

* To stop the dashboard, you can execute the following command:

```bash
sh devtron-install.bash stop
``` 

* To update the client, use the following command. It will stop the running dashboard and download the latest executable and open it in the browser.

```bash
sh devtron-install.bash upgrade
```



