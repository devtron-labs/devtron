## v1.3.0

## Bugs
- fix: panic fixes  (#6411)
- fix: auto pre-cd not getting triggered (if cd is manual) after webhook ci event is captured (#6407)
- fix: Base update (#6405)
- fix: devtron apps deployed via helm showing in helm apps list (#6397)
- fix: when a helm app is managed by argocd then skip argo app update when same name ext helm app is installed (#6389)
- fix: upgraded to go-git/v5 v5.13.0 (#6388)
- fix: fix for ci cd workflow fake success status and multiple post/pre cd success notifications (#6370)
- fix: fix for while cloning app or workflow when external secret is present in any pipeline then that pipeline is not cloned (#6368)
- fix: fix for while cloning app or workflow when external secret is present in any pipeline then that pipeline is not cloned (#6366)
- fix: buildx driver opt using k8s driver even if not enabled (#6349)
- fix: Auto trigger even if pre-cd fails (#6338)
- fix: role group listing fix (#6343)
- fix: sql injection fixes (#6334)
- fix: role group k8s permissions duplication (#6336)
## Enhancements
- feat: hibernate devtron app (#6356)
- feat: separate build worker status timeline (#6320)
- feat: send notification with chosen config (#6331)
- feat: sql script changes for operation audit  (#6342)
## Documentation
- doc: updated devtron user list in users.md (#6363)
## Others
- misc: pipeline listing in app Group (#6410)
- misc: Notif back com (#6398)
- sync: Main sync develop (#6396)
- sync: vendor update on release 29 (#6395)
- sync: Release candidate v0.29.0 (#6394)
- sync: Main sync develop (#6375)
- misc: common-lib change (#6347)
- sync: Release candidate v0.28.0 (#6340)
- misc: Fix typo: Cronjob → CronJob (#6335)
- sync: Main sync develop  (#6358)


