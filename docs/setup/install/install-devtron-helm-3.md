# Install using Helm3

### Install Devtron using Helm3

## Helm3

### Install with default configurations

This installation will use Minio for storing build logs and cache

```bash
kubectl create namespace devtroncd
helm add repo devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd --set secrets.POSTGRESQL_PASSWORD=change-me
```

### Install with AWS S3 Buckets

This installation will use AWS s3 buckets for storing build logs and cache

```bash
kubectl create namespace devtroncd
helm add repo devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=AWS \
--set configs.DEFAULT_CACHE_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CACHE_BUCKET_REGION=us-east-2 \
--set configs.DEFAULT_BUILD_LOGS_BUCKET=demo-s3-bucket \
--set configs.DEFAULT_CD_LOGS_BUCKET_REGION=us-east-2
```

### Install with Azure Blob Storage

This installation will use Azure Blob storage for storing build logs and cache

```text
kubectl create namespace devtroncd
helm add repo devtron https://helm.devtron.ai
helm install devtron devtron/devtron-operator --namespace devtroncd \
--set secrets.POSTGRESQL_PASSWORD=change-me \
--set configs.BLOB_STORAGE_PROVIDER=AZURE \
--set configs.AZURE_ACCOUNT_NAME=test-account \
--set configs.AZURE_BLOB_CONTAINER_CI_LOG=ci-log-container \
--set configs.AZURE_BLOB_CONTAINER_CI_CACHE=ci-cache-container
```

Copy and edit the `devtron/values.yaml` to configure your Devtron installation.

```bash
cp devtron/values.yaml devtron/install-values.yaml
vim devtron/install-values.yaml
```

For more details about configuration see the [helm chart configuration](install-devtron-helm-3.md#configuration) Once your configurations are ready, continue with following steps

```bash
#Create devtroncd namespace
kubectl create ns devtroncd
helm install devtron devtron/ --namespace devtroncd -f devtron/install-values.yaml
```

### Installation status

Run following command

```bash
kubectl -n devtroncd get installers installer-devtron -o jsonpath='{.status.sync.status}'
```

Once installation process is complete, above command will print `Applied` It may take around 30 minutes for installation to complete.

### Access devtron dashboard

#### Obtaining Dashboard URL when no BASE\_URL was provided

Obtaining Dashboard URL when no BASE\_URL was provided If you did not provide a \*\*BASE\_URL\*\* during install or have used the default installation, Devtron creates a loadbalancer for you on its own. Use the following command to get the dashboard url. \`\`\` kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}' \`\`\` You will get result something like below \`\`\` \[test2@server ~\]$ kubectl get svc -n devtroncd devtron-service -o jsonpath='{.status.loadBalancer.ingress}' \[map\[hostname:aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com\]\] \`\`\` The hostname mentioned here \( aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com \) is the Loadbalancer URL where you can access the Devtron dashboard. \*\*PS:\*\* You can also do a CNAME entry corresponding to your domain/subdomain to point to this Loadbalancer URL to access it at a custom domain. \| Host \| Type \| Points to \| \|----------:\|:------------\|:--------\| \| devtron.yourdomain.com \| CNAME \| aaff16e9760594a92afa0140dbfd99f7-305259315.us-east-1.elb.amazonaws.com \|

\#\#\#\# Obtaining Dashboard URL when a BASE\_URL was provided Devtron dashboard in now available at the \`BASE\_URL/dashboard\`, where \`BASE\_URL\` is same as provided in \`values.yaml\` in case of installation via helm chart OR provided in \`install/devtron-operator-configs.yaml\` in case of installation via kubectl. You can run following command to get dashboard \`\`\`bash scheme=\`kubectl -n devtroncd get cm devtron-operator-cm -o jsonpath='{.data.BASE\_URL\_SCHEME}'\` && url=\`kubectl -n devtroncd get cm devtron-operator-cm -o jsonpath='{.data.BASE\_URL}'\` && echo "$scheme://$url/dashboard" \`\`\` \*\*Please Note:\*\* URL should be pointing to the cluster on which you have installed the platform. For example if you have directed domain \`devtron.example.com\` to the cluster and ingress controller is listening on port \`32080\` then url will be \`devtron.example.com:32080\`

#### Devtron Admin credentials

For admin login use username:`admin` and for password run the following command.

```bash
kubectl -n devtroncd get secret devtron-secret -o jsonpath='{.data.ACD_PASSWORD}' | base64 -d
```

