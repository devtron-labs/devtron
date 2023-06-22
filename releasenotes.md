## v0.6.18-rc.3

## Bugs
- fix: Remove argocd server calls for Application resources (Manifest, Event, Logs, Terminal) (#3200)
- fix: disable api update if disabled from UI (#3487)
## Enhancements
## Documentation
## Others
- task: updated workflow name provision in app workflow api (#3526)


## v0.6.18-rc.2




## Bugs
- fix: Global secret data bug in CI workflow (#3504)
- fix: app workflow creation api panic and app metrics update issue (#3511)
- fix: HPA support for keda autoscaling and fixes for nameOverrides (#3522)
- fix : field validation and projects and chart repo fields (#3528)
- fix: acd app delete non cascade (#3506)
## Enhancements
- feat: releaseVersion in job of cron-job template 1.5.0 (#3358)
## Others
- task: added global policy db scripts (#3486)
- chore: replace `github.com/ghodss/yaml` with `sigs.k8s.io/yaml` (#3355)


## v0.6.18-rc.1



## Bugs
- Fix: argo workflow double encoding of secret value in template (#3496)
- fix: enable api for all users for module (#3495)
## Documentation
- docs: adding clusters on devtron using kubeconfig (#3494)
- docs: helm-charts permissions correction (#3484)


## v0.6.18-rc.0

## Bugs
- fix: shows commit details only on currently deployed image on deployment history (#3382)
- fix: sql script fix (#3474)
- fix: bulk-chart-deploy-fix (#3478)
- fix: image pull secret (#3488)
- fix: minor trivy fix (#3483)
- fix: argocd-cm empty data label nil pointer fix (#3443)
## Enhancements
- feat: validate kubeconfig and save clusters (#3148)
- feat: Cd refactoring migration (#3470)
- feat: Trivy Image Scanning (#3373)
- feat: added support for in app logging archival from CI runner (#3467)
## Documentation
- docs: updated doc for incluster (#3473)
- docs: ESO AWS secrets manager using clustersecretstore (#3453)
