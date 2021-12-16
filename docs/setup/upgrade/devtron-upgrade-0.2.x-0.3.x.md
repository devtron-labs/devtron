# Upgrading Devtron 0.2.x to 0.3.x

## Follow the required steps to update the Devtron version

#### STEP 1 

Delete the respective resources i.e, nats-operator , nats-streaming and nats-server using the following commands.

```bash
kubectl delete -f https://raw.githubusercontent.com/devtron-labs/devtron/v0.2.37/manifests/yamls/nats-operator.yaml
kubectl -n devtroncd delete -f https://raw.githubusercontent.com/devtron-labs/devtron/v0.2.37/manifests/yamls/nats-streaming.yaml
kubectl -n devtroncd delete -f https://raw.githubusercontent.com/devtron-labs/devtron/v0.2.37/manifests/yamls/nats-server.yaml
```

#### STEP 2

Verify the deletion of resources using the following commands. 

```bash
kubectl -n devtroncd get pods 
kubectl -n devtroncd get serviceaccount
kubectl -n devtroncd get clusterrole
```

#### STEP 3

Set `reSync: true` in the installer object, this will initiate upgrade of the entire Devtron stack, you can use the following command to do this.

```bash
kubectl patch -n devtroncd installer installer-devtron --type='json' -p='[{"op": "add", "path": "/spec/reSync", "value": true }]
```