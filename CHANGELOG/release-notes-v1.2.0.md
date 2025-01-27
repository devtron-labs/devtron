## v1.2.0



## Bugs
- fix: Application listing panic handling (#6199)
- fix: Copy image plugin panic handling (#6220)
- fix: Remove gaps between image scanning V1 & V2  (#6213)
- fix: Select Query errors in Postgres Pod Logs (#6238)
- fix: Graceful shutdown handling and logs flush (#6239)
- fix: Remove unused variable causing double register in prometheus (#6279)
## Enhancements
- feat: Ignore workflow cache config (#6218)
- feat: Build infra support (#6151)
- feat: Helm type pipeline app status (#6262)
- feat: Added labels to manual sync job (#6269)
## Others
- chore: Changes in CD trigger plugin condition v1.3.0 (#6233)
- misc: GetDockerRegistryIdForCiPipeline (#6231)
- chore: Merging user service with connection manager (#6225)
- chore: Argocd repository secret creation service refactoring (#6229)
- chore: Added sql migraiton for searchable key (#6252)
- misc: Adds the required env variables to set up devtron (#6254)
- misc: Added schme json for cronjobs (#6260)
- misc: Error handling when template is empty in ExtractVariables (#6261)
- misc: Nil check in envOverride.chart (#6263)
- misc: PG enable metrics (#6243)
- chore: Patch infra merge (#6274)
- chore: Argo common client (#6266)
- chore: Optimised Dockerfile by removing extra layers (#6293)
- chore: Updated ApplistingRestHandler oss/ent function (#6307)
- chore: Added env generator changes (#6302)
- chore: Deployment config methods add (#6297)
- misc: Deployment, reference and statefulset updated versions (#6284)
- chore: Scoped var oss/ent changes (#6319)
