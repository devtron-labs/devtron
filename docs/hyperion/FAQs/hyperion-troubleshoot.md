## Troubleshooting Guide

We always try to make your experience of using hyperion as smooth as possible but still if you face any issues, follow the troubleshooting guide given below or join our [discord channel](https://discord.gg/jsRG5qx2gp) if you couldn't find the solution for the issue you are facing.

### 1. Hyperion Installed but still helm apps are not visible on dashboard

To get helm apps on dashboard, it's important for migration jobs to be completed. To resolve this, check if Jobs are in `1/1 Completed` state by running the command:

```
kubectl get jobs -n devtroncd
```

If you see any of the jobs in `0/1 Completed` state then check if it's pod is still running using the following command:

```
kubectl get pods -n devtroncd
```

If the pods are in running state, then wait for them to complete and your helm apps should be visible on dashboard after that and if any of job's pod is in `CrashloopBackOff` state, then check the logs of that pod using:

```
kubectl logs -f <pod-name-in-crashloop> -n devtroncd -c <job-name>
```

Now, if you get something like `dirty db found` in the logs, then follow the steps given below and if not dirty db, then wait for the pod to automatically restart and complete the job so that helm apps are visible on dashboard.

#### Steps to follow in case you get dirty db found

1. Run this command to get postgresql password - `kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.PG_PASSWORD}' | base64 -d`
2. Copy the password you get and exec inside postgresql pod using `kubectl exec -it postgresql-postgresql-0 -n devtroncd -- sh`
3. Now when you are inside postgres, run the command to authenticate postgres user - `psql -U postgres` and enter the password that you got in step 1.
4. Terminate the connections to databases, delete them and then re-create using the commands given below
```
SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'orchestrator';
SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE pg_stat_activity.datname = 'casbin';
DROP database orchestrator;
DROP database casbin;
CREATE database orchestrator;
CREATE database casbin;
```
5. Download the migrator file and re-apply using the following commands:
```
kubectl delete -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/hyperion/migrator.yaml
kubectl apply -f https://raw.githubusercontent.com/devtron-labs/devtron/main/manifests/hyperion/migrator.yaml
```
After completing all the steps, you should see the helm apps on dashboard.