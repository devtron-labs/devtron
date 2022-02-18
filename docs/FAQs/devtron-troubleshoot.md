## Troubleshooting Guide

We always try to make your experience of using Devtron as smooth as possible but still if you face any issues, follow the troubleshooting guide given below or join our [discord channel](https://discord.gg/jsRG5qx2gp) if you couldn't find the solution for the issue you are facing.

#### 1. How to resolve unauthorized errors, while trying to save global configurations like hostname, GitOps etc. after successful devtron installation

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

#### 2. Not able to see deployment metrics on production environment or Not able to enable application-metrics or Not able to deploy the app after creating a configmap or secret with data-volume option enabled

Update the rollout crds to latest version, run the following command:
```
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/rollout.yaml -n devtroncd
```

#### 3. SSO Login not working even after entering correct SSO Credentials

Delete devtron pod once to reload the configurations using:
```
kubectl delete pod -n devtroncd -l app=devtron
```

#### 4. Logs are not Visible on UI while running the build and not even able to abort the same

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

#### 5. Grafana dashboards not visible in App Details page even after adding prometheus endpoint or Graphs showing error panel with id 2 not found

If the graphs are not visible check if prometheus is configured properly. Then go to Global Configurations > Clusters & Environments > Click on any environment for the cluster where you added prometheus endpoint and simply click `Update`.  
If the charts are still not visible, try visiting the url: <devtron-url>/grafana?orgId=2  
If you see `Not Found` on this page, then follow all the given steps or if the page is accessible and you are getting `panel with id 2 not found` then follow from step 6:  
1. Get grafana password using `kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.GRAFANA_PASSWORD}' | base64 -d`
2. `kubectl run --rm -it --image quay.io/devtron/k8s-utils:tutum-curl curl` Run this command and it will create a pod for using `curl`
3. Copy the following and change `grafana-password` with your password of grafana and change the value of `prometheusUrl` with your prometheus endpoint
```
cat << EOF
grafanaUrl = "http://admin:grafana-password@devtron-grafana.devtroncd/grafana"
prometheusUrl = "http://prometheus.example.com"

ORG_ID=$( curl -d '{"name":"devtron-metrics-view"}' -H "Content-Type: application/json" -X POST "${grafanaUrl}/api/orgs" )

echo $ORG_ID

curl -X POST "${grafanaUrl}/api/user/using/2";

curl -X PUT -H "Content-Type: application/json" -d '{"homeDashboardId":0,"theme":"light","timezone":"browser"}' "${grafanaUrl}/api/org/preferences";

curl "${grafanaUrl}/api/datasources" -H 'content-type: application/json' -H 'x-grafana-org-id: 2' --data '{"name":"Prometheus-devtron-demo","type":"prometheus","access":"proxy","isDefault":true}'

curl "${grafanaUrl}/api/datasources/2" -X PUT \
    -H 'content-type: application/json' \
    -H 'x-grafana-org-id: 2' \
    --data '{"id": 2 ,
    "orgId": 2,
    "name":"Prometheus-devtron-demo","type":"prometheus","access":"proxy",
    "url":${prometheusUrl},
    "basicAuth":false,"jsonData":{},"version":1}'
EOF
```
and run in the pod that we created above in step 2.
4. Now visit <devtron-url>/grafana?orgId=2 again and you'll see grafana login page. Login using username `admin` and password from step 1 and check if prometheus url is updated in datasources. If not, update it in the default datasource.
5. Now from devtron UI, update any of the environment again and it's datasource will be created automatically.
6. In Grafana UI you need to be logged in and Go to Dashboards > Manage then click `Import` and Import the given dashboards one by one.
```
https://grafana.com/api/dashboards/13322/revisions/4/download
https://grafana.com/api/dashboards/13320/revisions/4/download
https://grafana.com/api/dashboards/13325/revisions/4/download
https://grafana.com/api/dashboards/13321/revisions/6/download
```
After that, your issue should be resolved and you should be able to see all the graphs on UI.