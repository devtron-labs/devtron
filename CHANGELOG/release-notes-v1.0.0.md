## v1.0.0



We are extremely excited to announce Devtron's official release of Devtron 1.0. This marks an incredible milestone in our journey. Your valuable feedback and suggestions are what have helped us build the platform and constantly provide value to our users. This is just the beginning - we remain committed to innovation and will continue to deliver powerful features that enhance your DevOps experience based on community feedback and evolving industry needs.

## v1.0.0

## Bugs
- fix: Helm rollback in case of no gitops (#6005)
- fix: App detail deployed by (#6032)
- fix: Removed not null constraint for releaseOverride (#6044)
- fix: Showing inaccurate data in template/list api while fetching previous deployments list in template/list api, output shown was cartesian product of deployment_template_history and wfr (#6034)
- fix: Plugin getByName method error handling (#6016)
- fix: Config deployment history fix (#6003)
- fix: Force abort fix (#5990)
- fix: Helm deployment status  (#5996)
- fix: Deployment window FIXED type (#5986)
- fix: Migration seq fix (#5962)
- fix: Modified the query for User listing with filters (#5957)
- fix: Bulk deploy panic and docker tag handling (#5949)
- fix: Stage artifact logic  (#5913)
- fix: Permission access getting clubbed (#6045)
## Enhancements
- feat: Config diff enhancement (#5837)
- feat: Terminal role casbin policy (#5991)
- feat: Refactoring argo application service and common-lib constants (#5944)
## Others
- fix : Multiple param of filter in helm app list (#6013)
- chore: Adding refchart migration (#6007)
- chore: Cluster terminal images migration (#6006)
- misc: Removed deployment group validation when deleting CD pipelines (#5989)
- misc: Replaced != with <> for PostgreSQL compatibility (#5987)
- misc: isLatest field in plugin metadata tag changed to not_null (#5984)
- misc: Error handling while creating github repo in case of concurrent reqs (#5978)
- misc: Updated migration number (#5948)
- misc: Err handling for configmap access (#5938)
