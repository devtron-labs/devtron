# Argo CD application placement

`CD_NAMESPACE` selects the namespace of the Argo CD control-plane resources,
including newly created Applications and repository credentials. It defaults to
`devtroncd`, preserving bundled installations. Devtron installation settings
such as `ACD_NAMESPACE` can remain separate when Argo CD is installed elsewhere.

`ARGO_APPLICATION_PROJECT` selects the AppProject assigned to newly created
pipeline and chart-store Applications. Its default is `default`. Namespace and
project names are validated during configuration loading. Application workload
destinations, Git repositories and synchronization policy retain their existing
inputs. Changing these settings does not migrate or recreate existing Applications.

The platform owner supplies the AppProject's repository, destination and resource
constraints and the required Kubernetes/Argo permissions. Placement configuration
does not grant access, adopt an existing Git writer or provision another Argo CD.
