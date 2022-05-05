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

#### 6. Unable to Login Devtron dashboard even if the password is correct

If you are not able to login into Devtron dashboard even after giving the correct password, it is possible that the argocd token of previous session has been stored in the cookies and is not able to override the new token that is generated for the new session. If you are facing this issue, follow the steps below -

If using Firefox -
1. Goto login page of Devtron and open inspect.
2. Navigate to storage tab in inspect.
3. Click on url where Devtron has been installed under `Cookies` tab and you could see an argocd token with its value, something similar to below image.

![inspect-cookies](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron-troubleshooting/argocd-cookie.png)

4. Now right click on token, and click on `Delete All Session Cookies` option.

If using Chrome -
1. Goto login page of Devtron and open inspect.
2. Navigate to Application tab, and under `Storage` tab click on `Cookies`.
3. Click on url under `Cookie` and you would be able tto see an argocd token with its value, as shown in the image below.

![chrome-cookie](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron-troubleshooting/chrome-cookie.png)

4. Now right click on token and click on `delete` option.

If using Safari -
1. Goto Safari preferences >> Advanced options and check the show develop menu as shown in the image below.

![safari-preferences](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron-troubleshooting/safari-preferences.png)

2. Now goto login page of Devtron and press `option+command+I`. It will open inspect element.
3. Then navigate to `Storage`, click on `Cookies` and you would be able to see an argocd token with its value as shown in the image below.

![safari-cookie](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron-troubleshooting/safari-cookie.png)

4. Now right click on token and select `delete` option.

After clearing `Cookies`, try again to login, you should be able to login now.


#### 7. No charts found in Charts Discover Section

In the Devtron's Discover Chart section, if you are not able to see any charts available, goto `Global Configuration` >> `Chart Repositories` and click on `Refresh Chart` at the top-right as shown in the image below. After clicking the button, it might take 4-5mins to show all the charts in `Discover` section depending upon the chart repositories added.

![charts-not-found](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron-troubleshooting/refresh-charts.png)


#### 8. Not able to update cluster

In `Global Configurations` >> `Cluters & Environments`, if you try to update a cluster which has been already added in Devtron, you might get an error as `{"message":"Failed to update datasource. Reload new version and try again"}`. If you are facing such issue, please follow the following steps - 

1. Edit the changes you want to make in respective cluster
2. Click on save after making changes and you may get error message stated above.
3. Go to cluster where devtron has been installed and execute - `kubectl -ndevtroncd delete po -l app=devtron`
4. Now refresh the page and you should be able to save it.

[Note: If you already have created some environments in that cluster, it needs to be updated again]

#### 9. Postgresql is in crashloop with error - Failed to pull image
    
There may be some other pods also in crashloop as they are not able to connect to database. To resolve this issue, you can either [update devtron to latest version](https://docs.devtron.ai/devtron/setup/upgrade) or run the following commands to fix instantly on the same version you are using: 
```
kubectl patch -n devtroncd statefulset postgresql-postgresql -p '{"spec":{"template":{"spec":{"initContainers":[{"name":"init-chmod-data","image":"quay.io/devtron/minideb:latest"}],"containers":[{"name":"postgresql-postgresql","image":"quay.io/devtron/postgres:11.3.0-debian-9-r28"}]}}}}'
```
Then delete postgresql pod so that it can fetch the updated images:
```
kubectl delete pod -n devtroncd postgresql-postgresql-0
```
You can also delete other pods which are in crashloop after postgresql is up and running so that they can restart and connect to postgresql and Devtron will be up and running again in a few moments.

#### 10. Unable to fetch the latest commit and not able to trigger auto build.

To solve this, bounce the git-sensor-0 pod.
```
kubectl delete pod -n devtroncd git-sensor-0
```
#### 11. If you have restricted devtron-service to be accessible on certain IPs only and SSO login isn’t working

Whitelist the NAT-gateway IPs of the cluster (There can be multiple NAT-gateways if your cluster is multi-AZ)

#### 12. If CPU metrics are not showing but memory metrics are visible in graphs.

Do the following:-

1. Go to Grafana and Login with the credentials.
2. Edit the CPU graphs and remove `image!=””` from the query.
3. Save the dashboard.

CPU metrics should start showing up in a while.

#### 13. If user not able to upload a file more than specific size. 

`Please use below annotation in ingress`
```
nginx.ingress.kubernetes.io/proxy-body-size: 100m
```
`Note:- `Where m is is MiB.

#### 14. If AWS Load balancer controller is unable to provision ALB and getting message in alb controller as unauthorized, attach these IAM policy to  the nodegroup IAM Role.

[IAM policy](https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.3.1/docs/install/iam_policy.json)

#### 15. When app metrics is not coming on grafana and devtron dashboard, set the value of the following parameter as false in kube prometheus stack values.

```
serviceMonitorSelectorNilUsesHelmValues: false
```
#### 16. Unable to deploy metrics-server using chart on devtron

To solve

Disable certificate validation by passing `--kubelet-insecure-tls` argument to metrics server chart.

#### 17. Unable to delete a database from postgres
`Description of issue`

ERROR: database `<db-name>` is being accessed by other users

DETAIL: There is 1 other session using the database.

You have to terminate the connections to the database first, for that you can use the command.
```
SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'TARGET_DB';
```
Then run the command to delete database - `drop databases <db-name>`

#### 18. Unable to login with the auth password (argocd server)

`Debug`

Run the command for Admin Credentials and use it for login in dashboard

`kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d`

If you are getting an error message of  “invalid username or password”, follow the solution to solve it.

`Solution`

Run `kubectl get secret -n devtroncd` and then edit the `argocd-secret`, remove both the admin.password lines.

Run `kubectl delete po your-argocd-server-pod -n devtroncd`, it will create a new pod after deletion and reset your admin password. Re-run the command for admin credentials again to get the new password.

#### 19. After installing Devtron using helm, getting the admin password does not work.(if using windows)

`Debug`

'base64' is not recognized as an internal or external command, operable program or batch file.

`Solution`

The first way  to debug is either install base64 encode and decode into your windows machine and use the appropriate cmd to get the admin password.

The other way is to get the password in the encoded form using the cmd

`kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}'`, further decode it into plaintext using an online [encoder decoder](https://www.base64decode.org/).



