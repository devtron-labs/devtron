## 1. How to resolve unauthorized error while trying to save global configurations like hostname, gitops etc. after successful devtron installation
<br>
This occurs most of the time because any one or multiple jobs get failed during installation. To resolve this, you need to first check which are the jobs that have failed. Follow these steps :-

- Run the following command and check which are the jobs with 0/1 completions:
```bash
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
- If your problem is still not resolved, you can post your query in our [discord](https://discord.gg/jsRG5qx2gp) channel


## 2. What to do if devtron dashboard is not accessible on browser even after successful completion of all the jobs and all pods are in running mode


- For this, you need to check if nats-cluster is created or not, you can check it using the following command:
```bash
kubectl get natscluster -n devtroncd
```
- You should see a natscluster with the name devtron-nats and if not, run the given command:
```bash
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/yamls/nats-server.yaml -n devtroncd
```
- Wait till all the nats pods are created and the pods are in running condition. After that delete devtron and dashboard pods once and then you should be able to access the devtron dashboard without any issues.
- If your problem is still not resolved, you can post your query in our [discord](https://discord.gg/jsRG5qx2gp) channel