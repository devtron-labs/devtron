# An externally managed Argo CD account

Set `CD_HOST` and `CD_PORT` to the TLS endpoint of the existing Argo CD server.
Set `CD_TOKEN_FILE` to a read-only mounted file containing its API token. This
disables Devtron's automatic Argo CD account/password/token provisioning and
does not require access to the Argo CD server's credential or signing Secrets.
The token issuer owns the account's permissions, expiration and renewal.

Optionally set `CD_TLS_CA_FILE` to a read-only PEM CA bundle. External token mode
uses the system trust store when this is absent; certificate and hostname
verification remain enabled. An invalid configured CA fails initialization.
The CA bundle is loaded during initialization; rotate it through the normal
workload rollout. The token file is read for every RPC, so mounted token updates
take effect on existing connections. Missing or empty token files fail the RPC
without provisioning another account or sending an anonymous request.

With both file settings unset, the existing bundled-Argo account and TLS path
is retained. Kubernetes permissions for application/repository resources remain
separate from gRPC credentials and must be configured for the intended namespace.
