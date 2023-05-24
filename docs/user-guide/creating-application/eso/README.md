# External Secret Operator (ESO)

> **Prerequisites:** Chart version should be > 4.14.0

External Secrets Operator is a Kubernetes operator that integrates external secret management systems like AWS Secrets Manager, HashiCorp Vault, Google Secrets Manager, Azure Key Vault and many more. The operator reads information from external APIs and automatically injects the values into a Kubernetes Secret.

## Install External Secret Operator

Before creating any external secrets on Devtron, `External Secret Operator` must be installed on the target cluster.  `External Secret Operator` allows you to use external secret management systems (e.g., AWS Secrets Manager, Hashicorp Vault, Azure Secrets Manager,  Google Secrets Manager etc.) to securely inject secrets in Kubernetes.

You can install `External Secrets Operator` using charts store:

1. Go to charts store.
2. Search chart with name `external-secrets`.

![External secrets chart](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/secrets/external-secret.png)

3. If you don't find any chart with this name i.e `external-secrets`, add chart repository using repository url  ` https://charts.external-secrets.io`. Please follow this [documentation](https://docs.devtron.ai/getting-started/global-configurations/chart-repo#add-chart-repository) for adding chart repository.
4. Deploy the chart.