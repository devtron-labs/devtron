# Argo CD application placement

`ACD_NAMESPACE` is the existing Argo CD namespace setting used by chart-store
Applications, repository resources and default Application lookups. Pipeline
Application creation now honors that same setting instead of always using
`devtroncd`. The default remains `devtroncd`. The connection/account bootstrap
setting `CD_NAMESPACE` retains its existing behavior and is not a replacement for
`ACD_NAMESPACE`.

`ARGO_APPLICATION_PROJECT` selects the AppProject assigned to newly created
pipeline and chart-store Applications. Its default is `default`. Namespace and
project names are validated during deployment configuration loading. Application
workload destinations, Git repositories and synchronization policy retain their
existing inputs. Changing these settings does not migrate or recreate existing
Applications. Existing repository and Application lookup configuration is unchanged.

The platform owner supplies the AppProject's repository, destination and resource
constraints and required Kubernetes/Argo permissions. Placement configuration
does not grant access, adopt an existing Git writer or provision another Argo CD.

New pipeline deployment metadata records the same configured Application namespace.
Discovery uses recorded placement for managed and linked pipelines, preserving
legacy pipeline placement when no metadata is available. Chart-store discovery
and duplicate-link checks retain the configured Argo CD namespace. Changing the
sync mode preserves the fetched Application metadata, project and sources.
