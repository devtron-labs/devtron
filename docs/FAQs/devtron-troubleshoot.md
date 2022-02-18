## Troubleshooting Guide

We always try to make your experience of using Devtron as smooth as possible but still if you face any issues, follow the troubleshooting guide given below or join our [discord channel](https://discord.gg/jsRG5qx2gp) if you couldn't find the solution for the issue you are facing.

### 1. How to resolve unauthorized error's, while trying to save global configurations like hostname, GitOps etc. after successful devtron installation

This occurs most of the time because any one or multiple jobs get failed during installation. To resolve this, you'll need to first check which jobs have failed. Follow these steps:

- Run the following command and check which are the jobs with 0/1 completions:
```
kubectl get jobs -n devtroncd
```
- Note down or remember the names of jobs with 0/1 completions and check if their pods are in running state still or not by running the command:
kubectl get pods -n devtroncd
- If they are in running condition, please wait for the jobs to be completed as it may be due to internet issue and if not in running condition, then delete those incomplete jobs using:
kubectl delete jobs <job1-name> <job2-name> -n devtroncd
- Now download migrator.yaml file from our github repository using the command:
wget https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/migrator.yaml
- Now edit the file you downloaded in step 3 and remove the postgresql-migrator secret resource creation and then apply the yaml file using the command:
kubectl apply -f migrator.yaml -n devtroncd
- It will re-create the failed jobs and you’ll see their pods created again. Just wait for a few minutes until the jobs gets completed then you are good to go. You should be able to save your global configurations now.

### 2. Not able to see deployment metrics on production environment or Not able to enable application-metrics or Not able to deploy the app after creating a configmap or secret with data-volume option enabled

Update the rollout crds to latest version, run the following command:
```
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/rollout.yaml -n devtroncd
```

### 3. SSO Login not working even after entering correct SSO Credentials

Delete devtron pod once to reload the configurations using:
```
kubectl delete pod -n devtroncd -l app=devtron
```

### 4. Logs are not Visible on UI while running the build and not even able to abort the same

Check if the pods are being created when you start a new build, run the command and look if a new pod is created when you started the build:
```
kubectl get pods -n devtron-ci
```
If yes, delete kubewatch and devtron pod so that kubewatch can restart and start sharing the logs again:
```
kubectl delete pod -n devtroncd -l app=devtron; kubectl delete pod -n devtroncd -l app=kubewatch
```
Wait for 5 minutes and then trigger a new build again, if still not resolved then run the following commands one by one
```
kubectl delete pod -n devtroncd devtron-nats-0
kubectl delete pod -n devtroncd devtron-stan-0
kubectl delete pod -n devtroncd -l app=devtron
kubectl delete pod -n devtroncd -l app=kubewatch
```
Again wait for 5 minutes and your issue should be resolved