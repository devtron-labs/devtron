## v0.6.29



## v0.6.29



## v0.6.29



## v0.6.29



## v0.6.29



## v0.6.29



## v0.6.29



## Bugs
- fix: removed unused dependency (#4916)
- fix: release not found alert (#4928)
- fix: link external helm app to chart store (#4919)
- fix: Corrected the error of fetching app name and pipeline name from CI_CD_EVENT (#4911)
- fix:job name support added in input variables (#4863)
- fix:duplicate workflow name fixed (#4841)
- fix: deprecate latest in app store app version (#4896)
- fix: rbac fix in case of project in upper case (#4840)
- fix: use join cookie logic to extract token (#4892)
- fix: chart sync job optimisation (#4631)
- fix: depandabot version upgrade (#4792)
- fix: injected app-serveice dependency into DeployedApplicationEventProcessor service (#4875)
- fix: update argo app repo url in patch (#4876)
- fix: removed code for gitops repo migration in devtron apps (#4838)
- fix: return nil,err on git material fetch error (#4857)
- fix: ns not found in case ips is being injected in cluster (in global config) (#4844)
- fix: helm deployments stucked in queued for devtron apps (#4842)
- fix: panic on re-trigger ci if pod deleted on BuildTriggerObject (#4826)
- fix: hide ldap creds in in get req (#4788)
- fix: 5xx 4.0 iter (#4620)
- fix:  rolefilters correction with all applications and particular application selected. (#4820)
- fix: extra check added for mono-repo migraiton (#4764)
- fix: trim space from git repo Url on create and update material (#4787)
## Enhancements
- perf: memory optimizations and prom metrics on terminal session exposed (#4909)
- feat: Added the Devtron CI Trigger Plugin (#4908)
- feat:MailMaster Plugin v1.0 (#4825)
- feat: Add support for git hash in the app and app group overview section _revised (#4836)
- feat:Github Pull Request Closer (#4833)
- feat: Added Apply job in k8s plugin (#4828)
- feat: Branch Divergence Checker Plugin (#4806)
- feat: added request method in audit logger (#4817)
- feat: active inactive user phase 3 (#4649)
## Documentation
- doc: Added EnvVariablesFromFieldPath in Deployment Template (#4852)
## Others
- chore: migration support for notification db changes (#4906)
- chore: App Store refactoring v4 (#4647)
- chore: Resource mapping refactoring (#4813)
- misc: Refactoring Pagerduty Issue Calculator script (#4856)
- misc: Update github_pagerduty_score_calculation.yml (#4853)
- misc: Update pager-duty.yaml (#4850)
- misc: Pagerduty issue template change (#4796)
- chore: Image scanning refactoring (#4802)
- chore: refactoring v4 (#4775)


